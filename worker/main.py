# Copyright 2024 Google, LLC.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""Worker module for Litmus, executing tests and storing results."""

import json
import os
import re
from datetime import datetime
from uuid import uuid4

import requests
from google.cloud import firestore, logging, storage

from util.assess import (
    ask_llm_against_golden,
    ask_llm_for_action,
    is_mission_done,
    evaluate_mission,
)

# Setup logging
# Instantiates a logging client
logging_client = logging.Client()

# Define log names
CORE_LOG_NAME = "litmus-core-log"
WORKER_LOG_NAME = "litmus-worker-log"

# Selects the logs to write to
core_logger = logging_client.logger(CORE_LOG_NAME)
worker_logger = logging_client.logger(WORKER_LOG_NAME)

# Writes a log entry indicating the worker is starting
worker_logger.log_text("### Litmus-worker starting ###")

# Initialize Storage client
storage_client = storage.Client()

# Get the files bucket name and prefix from environment variable
files_bucket_name = os.environ.get("FILES_BUCKET")
files_prefix = os.environ.get("FILES_PREFIX", "")  # Default to no prefix
files_bucket = storage_client.bucket(files_bucket_name)


def execute_request(request_data):
    """Executes a given HTTP request and returns the response.

    Args:
        request_data: A dictionary containing the request data, including:
            - url (str): The URL to send the request to.
            - method (str, optional): The HTTP method (default: 'POST').
            - body (dict, optional): The request body.
            - headers (dict, optional): The request headers.
            - tracing_id (str, optional): A unique ID for tracing the request.

    Returns:
        tuple: A tuple containing:
            - dict: The JSON response from the server if successful, or an error message if the request fails.
            - int: The HTTP status code (or 0 if an exception occurs).
    """

    url = request_data.get("url")
    method = request_data.get("method", "POST")
    body = request_data.get("body")
    headers = request_data.get("headers")
    tracing_id = request_data.get("tracing_id")

    # Add tracing ID to headers for tracking
    if tracing_id:
        headers["X-Litmus-Request"] = tracing_id

    # Process file references in the request body
    body = process_file_references(body)

    start_time = datetime.utcnow()
    status_code = 0  # Default status code in case of exceptions

    try:
        # Send the HTTP request based on the specified method
        if method == "POST":
            response = requests.post(url, json=body, headers=headers)
        elif method == "GET":
            response = requests.get(url, headers=headers)
        elif method == "PUT":
            response = requests.put(url, json=body, headers=headers)
        elif method == "DELETE":
            response = requests.delete(url, headers=headers)
        else:
            raise ValueError(f"Unsupported HTTP method: {method}")

        status_code = response.status_code
        response.raise_for_status()  # Raise an exception for bad status codes

        end_time = datetime.utcnow()
        log_request_and_response(
            request_data, response, start_time, end_time
        )  # Log the request and response

        return response.json(), status_code
    except requests.exceptions.RequestException as e:
        end_time = datetime.utcnow()
        log_request_and_response(
            request_data, None, start_time, end_time, error=str(e)
        )  # Log the error
        return {"status": "Failed", "error": str(e)}, status_code
    except Exception as e:
        end_time = datetime.utcnow()
        log_request_and_response(
            request_data, None, start_time, end_time, error=str(e)
        )  # Log the error
        return {"status": "Failed", "error": str(e)}, status_code


def process_file_references(data):
    """Replaces file references in the data with the content of the referenced files.

    Args:
        data: The request data (could be a string, dictionary, or list).

    Returns:
        The processed data with file references replaced.
    """
    if isinstance(data, str):
        return replace_file_reference_in_string(data)
    elif isinstance(data, dict):
        for key, value in data.items():
            data[key] = process_file_references(value)
        return data
    elif isinstance(data, list):
        return [process_file_references(item) for item in data]
    else:
        return data


def replace_file_reference_in_string(text):
    """
    Replaces file references in a string with the content of the referenced file.

    Args:
        text (str): The text that may contain file references.

    Returns:
        str: The text with file references replaced.
    """

    pattern = r"\[FILE:\s*(.+?)\]"
    matches = re.findall(pattern, text)

    for match in matches:
        file_content = read_file_from_gcs({files_prefix}{match})

        text = text.replace(f"[FILE: {match}]", file_content)

    return text


def read_file_from_gcs(file):
    """Reads the content of a file from Google Cloud Storage.

    Args:
        gcs_path: The full GCS path to the file (e.g., "gs://my-bucket/my-file.txt").

    Returns:
        str: The content of the file, or an error message if reading fails.
    """

    try:
        blob = files_bucket.blob(file)  # Remove "gs://" prefix
        return blob.download_as_text()
    except Exception as e:
        worker_logger.log_text(
            f"Error reading file from GCS: {file}, {str(e)}", severity="ERROR"
        )
        return f"Error reading file: {file}"


def execute_test_mission(run_data, test_case, test_case_ref, tracing_id):
    """Executes a test mission, interacting with the LLM iteratively.

    Args:
        run_data (dict): Data for the current test run.
        test_case (dict): Data for the current test case.
        test_case_ref (firestore.DocumentReference): Reference to the test case document.
        tracing_id (str): Unique ID for tracing requests.

    Returns:
        dict: The test mission result, including status, conversation history, assessment, and payloads.
    """

    mission_duration = run_data.get("mission_duration")
    mission_description = test_case.get("mission")
    conversation_history = []
    request_response_history = []
    test_result = {}

    for turn in range(mission_duration):
        worker_logger.log_text(f"Mission turn: {turn+1}/{mission_duration}")

        try:
            # Ask LLM for the next action based on conversation history
            llm_action = ask_llm_for_action(mission_description, conversation_history)

            # Handle cases where LLM doesn't provide a valid action
            if not llm_action or "request" not in llm_action:
                worker_logger.log_text(
                    f"Error: LLM returned invalid action: {llm_action}",
                    severity="ERROR",
                )
                test_result = {
                    "status": "Error",
                    "error": f"Invalid LLM action on turn {turn+1}",
                }
                break

            request_data = test_case.get("request")
            request_data["tracing_id"] = tracing_id
            json_string = json.dumps(request_data)
            json_string = json_string.replace(f"{{query}}", str(llm_action["request"]))
            request_data = json.loads(json_string)
            conversation_history.append(
                {"role": "user", "content": llm_action["request"]}
            )

            # Execute the request suggested by the LLM
            actual_response, status_code = execute_request(request_data)

            # Store request - response pair with status code
            request_response_history.append(
                {
                    "request": request_data,
                    "response": actual_response,
                    "status_code": status_code,
                }
            )

            # Handle potential errors in the API response
            if "status" in actual_response and actual_response["status"] == "Failed":
                worker_logger.log_text(
                    f"API request failed on turn {turn+1}: {actual_response['error']}",
                    severity="ERROR",
                )
                test_result = {
                    "status": "Failed",
                    "error": f"API request failed: {actual_response['error']}",
                }
                break

            # Add the response to the conversation history
            actual_filtered_response = filter_json(
                actual_response, run_data.get("template_output_field")
            )
            try:
                conversation_history.append(
                    {
                        "role": "assistant",
                        "content": actual_filtered_response[
                            run_data.get("template_output_field")
                        ],
                    }
                )
            except:
                conversation_history.append(
                    {
                        "role": "assistant",
                        "content": actual_filtered_response,
                    }
                )

            # Check if the mission is done
            if is_mission_done(mission_description, conversation_history):
                worker_logger.log_text(
                    f"Mission completed successfully on turn {turn+1}"
                )
                test_result = {
                    "status": "Passed",
                    "conversation_history": conversation_history,
                }
                break

        except Exception as e:
            worker_logger.log_text(
                f"Error in mission execution on turn {turn+1}: {str(e)}",
                severity="ERROR",
            )
            test_result = {"status": "Failed", "error": str(e)}
            break

    # Evaluate the entire mission using the LLM after the loop
    final_assessment = evaluate_mission(
        mission_description,
        conversation_history,
        test_case.get("golden_response"),
        run_data.get("template_llm_prompt"),
    )

    result = {}

    result["turns"] = len(conversation_history) / 2
    result["conversation"] = conversation_history
    result["assessment"] = final_assessment
    result["payloads"] = request_response_history
    result["result"] = test_result
    try:
        result["status"] = final_assessment["overall_success"]
    except:
        result["status"] = "Failed"

    return result


def execute_test_run(run_data, test_case, tracing_id):
    """Executes a single test run, comparing the response with the golden response.

    Args:
        run_data (dict): Data for the current test run.
        test_case (dict): Data for the current test case.
        tracing_id (str): Unique ID for tracing requests.

    Returns:
        dict: The test result, including status, response, and optional assessment.
    """

    request_data = test_case.get("request")
    request_data["tracing_id"] = tracing_id
    golden_response = test_case.get("golden_response")

    try:
        # Execute the main request
        actual_response, status_code = execute_request(request_data)

        # Handle potential errors in the API response
        if "status" in actual_response and actual_response["status"] == "Failed":
            return {
                "status": "Failed",
                "error": f"API request failed: {actual_response['error']}",
                "status_code": status_code,
            }

        output_field = run_data.get("template_output_field")
        template_llm_prompt = run_data.get("template_llm_prompt")
        actual_filtered_response = filter_json(
            actual_response, run_data.get("template_output_field")
        )

        # Compare with golden response if available
        if golden_response and actual_filtered_response:
            try:
                # Assess the actual response against the golden response using an LLM
                llm_assessment = ask_llm_against_golden(
                    statement=actual_filtered_response.get(output_field),
                    golden=golden_response,
                    prompt=template_llm_prompt,
                )

                # Evaluate LLM assessment results
                if llm_assessment and "similarity" in llm_assessment:
                    if llm_assessment.get("similarity") > 0.5:
                        test_result = {
                            "status": "Passed",
                            "response": actual_response,
                            "assessment": llm_assessment,
                            "status_code": status_code,
                        }
                    else:
                        test_result = {
                            "status": "Failed",
                            "expected": golden_response,
                            "response": actual_response,
                            "assessment": llm_assessment,
                            "status_code": status_code,
                        }
                else:
                    # Handle invalid LLM assessment
                    test_result = {
                        "status": "Error",
                        "response": actual_response,
                        "error": "LLM assessment returned an invalid response",
                        "status_code": status_code,
                    }

            except Exception as e:
                # Log errors from the LLM assessment
                worker_logger.log_text(
                    f"Error in ask_llm_against_golden: {str(e)}",
                    severity="ERROR",
                )
                test_result = {
                    "status": "Error",
                    "response": actual_response,
                    "error": f"Error during LLM assessment: {str(e)}",
                    "status_code": status_code,
                }
        elif actual_filtered_response:
            # Handle cases where no golden response is provided
            test_result = {
                "status": "Passed",
                "response": actual_response,
                "note": "No golden response available",
                "status_code": status_code,
            }
        else:
            test_result = {
                "status": "Failed",
                "response": actual_response,
                "note": "No response available",
                "status_code": status_code,
            }

    except Exception as e:
        test_result = {"status": "Failed", "error": str(e), "status_code": status_code}

    return test_result


def execute_tests_and_store_results(run_id, template_id):
    """Executes tests from a template and stores results, updating progress.

    Args:
        run_id (str): The ID of the test run.
        template_id (str): The ID of the test template.
    """

    db = firestore.Client()
    run_ref = db.collection("test_runs").document(run_id)
    run_data = run_ref.get().to_dict()

    if not run_data:
        worker_logger.log_text(f"Error: Run ID '{run_id}' not found.")
        return

    if not template_id:
        worker_logger.log_text(f"Error: Template ID not found for run '{run_id}'")
        return

    # Retrieve test cases from Firestore
    test_cases_ref = db.collection(f"test_cases_{run_id}")
    test_cases = [doc.to_dict() for doc in test_cases_ref.stream()]

    # Update run status to "Running"
    run_ref.update({"status": "Running"})
    num_tests = len(test_cases)
    num_completed = 0
    worker_logger.log_text(f"Running {num_tests} tests")

    # Iterate through each test case
    for i, test_case in enumerate(test_cases):
        tracing_id = str(uuid4())  # Generate a unique tracing ID

        try:
            # Execute pre-request hook if defined
            if test_case.get("pre_request"):
                test_case["pre_request"]["tracing_id"] = tracing_id
                execute_request(test_case["pre_request"])

            # If "Test Mission" - execute_test_mission
            if run_data.get("template_type") == "Test Mission":
                test_result = execute_test_mission(
                    run_data,
                    test_case,
                    test_cases_ref.document(f"test_case_{i+1}"),
                    tracing_id,
                )
            else:  # "Test Run"
                test_result = execute_test_run(run_data, test_case, tracing_id)

            # Execute post-request hook if defined
            if test_case.get("post_request"):
                test_case["post_request"]["tracing_id"] = tracing_id
                execute_request(test_case["post_request"])

        except Exception as e:
            # Log and store any errors that occur during test execution
            worker_logger.log_text(
                f"Error executing test case {i+1}: {str(e)}", severity="ERROR"
            )
            test_result = {"status": "Failed", "error": str(e)}

        # Store the test result in Firestore
        test_case_ref = db.collection(f"test_cases_{run_id}").document(
            f"test_case_{i+1}"
        )
        test_case_ref.update({"result": test_result})
        test_case_ref.update({"tracing_id": tracing_id})

        num_completed += 1

        # Update the progress of the test run
        run_ref.update({"progress": f"{num_completed}/{num_tests}"})

    end_time = datetime.utcnow()

    # Update run status to "Completed"
    run_ref.update({"status": "Completed", "end_time": end_time})
    worker_logger.log_text(f"Running tests completed")


def filter_json(data, filter_pathx):
    """Filters a JSON structure based on the given path.

    Args:
        data (dict): The JSON data to filter.
        filter_pathx (str): A comma-separated string representing the paths to the desired values.
            Uses dot notation (e.g., "key1.key2.key3") and supports array indexing (e.g., "key1.key2[0].key3").

    Returns:
        dict: A dictionary containing the filtered values.
    """

    if not filter_pathx:
        return data

    filter_paths = filter_pathx.split(",")
    new_data = {}

    for filter_path in filter_paths:
        keys = filter_path.split(".")
        current_data = data

        # Traverse the JSON structure based on the keys in the filter path
        for key in keys:
            if current_data is not None:
                # Handle array indexing
                if "[" in key and "]" in key:
                    key, index = key.split("[", 1)[0], int(key.split("[", 1)[1][:-1])
                    if 0 <= index < len(current_data) and key in current_data:
                        current_data = current_data[key][index]
                    else:
                        current_data = None
                        break
                # Access the nested data using the key
                elif key in current_data:
                    current_data = current_data[key]
                else:
                    # Key not found, move to the next filter path
                    current_data = None
                    break

        # If the traversal is successful, add the filtered value to new_data
        if current_data is not None:
            new_data[filter_path] = current_data

    return new_data


def log_request_and_response(request_data, response, start_time, end_time, error=None):
    """Logs details of an HTTP request and its response.

    Args:
        request_data (dict): The data of the request.
        response (requests.Response, optional): The response object.
        start_time (datetime): The time the request was sent.
        end_time (datetime): The time the response was received.
        error (str, optional): An error message, if any.
    """

    request_log = {
        "id": str(uuid4()),
        "tracingID": request_data.get("tracing_id"),
        "timestamp": start_time.isoformat(),
        "method": request_data.get("method"),
        "requestURI": request_data.get("url"),
        "requestHeaders": request_data.get("headers"),
        "requestBody": request_data.get("body"),
        "requestSize": len(json.dumps(request_data.get("body"))),
        "latency": (end_time - start_time).total_seconds()
        * 1000,  # Latency in milliseconds
    }

    if response:
        request_log["responseStatus"] = response.status_code
        try:
            request_log["responseBody"] = response.json()
        except ValueError:
            request_log["responseBody"] = response.text
        request_log["responseSize"] = len(response.content)
    if error:
        request_log["error"] = error

    core_logger.log_struct(request_log)


if __name__ == "__main__":
    run_id = os.environ.get("RUN_ID")
    template_id = os.environ.get("TEMPLATE_ID")
    if not run_id or not template_id:
        raise ValueError("RUN_ID and TEMPLATE_ID environment variables must be set")

    execute_tests_and_store_results(run_id, template_id)
