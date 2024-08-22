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

import os
import requests
import json
from google.cloud import firestore
from datetime import datetime
from util.assess import ask_llm_against_golden
from google.cloud import logging
from uuid import uuid4

# setup logging
# Instantiates a client
logging_client = logging.Client()

# Define log names
core_log_name = "litmus-core-log"
worker_log_name = "litmus-worker-log"

# Selects the logs to write to
core_logger = logging_client.logger(core_log_name)
worker_logger = logging_client.logger(worker_log_name)

# Writes the log entry
worker_logger.log_text("### Litmus-worker starting ###")


def execute_request(request_data):
    """Executes a given request and returns the response."""
    url = request_data.get("url")
    method = request_data.get("method", "POST")  # Default to POST
    body = request_data.get("body")
    headers = request_data.get("headers")
    tracing_id = request_data.get("tracing_id")  # Retrieve tracing ID

    # Add tracing ID to headers if it exists
    if tracing_id:
        headers["X-Litmus-Request"] = tracing_id

    start_time = datetime.utcnow()  # Capture request start time

    try:
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

        response.raise_for_status()

        end_time = datetime.utcnow()  # Capture request end time
        log_request_and_response(
            request_data, response, start_time, end_time
        )  # Log request and response

        return response.json()
    except requests.exceptions.RequestException as e:
        end_time = datetime.utcnow()  # Capture request end time
        log_request_and_response(
            request_data, None, start_time, end_time, error=str(e)
        )  # Log error
        return {"status": "Failed", "error": str(e)}


def execute_tests_and_store_results(run_id, template_id):
    """Executes tests from a template and stores results, updating progress."""
    db = firestore.Client()
    run_ref = db.collection("test_runs").document(run_id)
    run_data = run_ref.get().to_dict()

    if not run_data:
        worker_logger.log_text(f"Error: Run ID '{run_id}' not found.")
        return

    if not template_id:
        worker_logger.log_text(f"Error: Template ID not found for run '{run_id}'")
        return

    # Get test cases from the subcollection
    test_cases_ref = db.collection(f"test_cases_{run_id}")
    test_cases = [doc.to_dict() for doc in test_cases_ref.stream()]

    # Update run status to "Running"
    run_ref.update({"status": "Running"})
    num_tests = len(test_cases)
    num_completed = 0
    worker_logger.log_text(f"Running {num_tests} tests")

    for i, test_case in enumerate(test_cases):
        tracing_id = str(uuid4())  # Generate a unique tracing ID for each test case

        # Execute pre-request (if available)
        if test_case.get("pre_request"):
            test_case["pre_request"]["tracing_id"] = tracing_id  # Add tracing ID
            execute_request(test_case["pre_request"])

        request_data = test_case.get("request")
        request_data["tracing_id"] = tracing_id  # Add tracing ID
        golden_response = test_case.get("golden_response")

        try:
            actual_response = execute_request(request_data)

            # Compare with golden response
            if golden_response:
                # Exception handling for ask_llm_against_golden
                try:
                    llm_assessment = ask_llm_against_golden(
                        statement=actual_response.get("output").get("text"),
                        golden=golden_response.get("text"),
                    )

                    # Check if llm_assessment is valid
                    if llm_assessment and "similarity" in llm_assessment:
                        if llm_assessment.get("similarity") > 0.5:
                            test_result = {
                                "status": "Passed",
                                "response": actual_response,
                                "assessment": llm_assessment,
                            }
                        else:
                            test_result = {
                                "status": "Failed",
                                "expected": golden_response,
                                "response": actual_response,
                                "assessment": llm_assessment,
                            }
                    else:
                        # Handle invalid llm_assessment
                        test_result = {
                            "status": "Error",
                            "response": actual_response,
                            "error": "LLM assessment returned an invalid response",
                        }

                except Exception as e:
                    # Log the specific error from ask_llm_against_golden
                    worker_logger.log_text(
                        f"Error in ask_llm_against_golden: {str(e)}", severity="ERROR"
                    )
                    test_result = {
                        "status": "Error",
                        "response": actual_response,
                        "error": f"Error during LLM assessment: {str(e)}",
                    }
            else:
                # Handle case where golden response is missing
                test_result = {
                    "status": "Passed",
                    "response": actual_response,
                    "note": "No golden response available",
                }

        except requests.exceptions.RequestException as e:
            test_result = {"status": "Failed", "error": str(e)}

        # Execute post-request (if available)
        if test_case.get("post_request"):
            test_case["post_request"]["tracing_id"] = tracing_id  # Add tracing ID
            execute_request(test_case["post_request"])

        # Store test result in Firestore
        test_case_ref = db.collection(f"test_cases_{run_id}").document(
            f"test_case_{i+1}"
        )
        test_case_ref.update({"result": test_result})

        num_completed += 1

        # Update run progress
        run_ref.update({"progress": f"{num_completed}/{num_tests}"})

    end_time = datetime.utcnow()
    # Update run status to "Completed"
    run_ref.update({"status": "Completed", "end_time": end_time})
    worker_logger.log_text(f"Running tests completed")


def log_request_and_response(request_data, response, start_time, end_time, error=None):
    """Logs the request and response details to the core logger."""

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

    core_logger.log_struct(request_log)  # Log the structured data to the core logger


if __name__ == "__main__":
    run_id = os.environ.get("RUN_ID")
    template_id = os.environ.get("TEMPLATE_ID")
    if not run_id or not template_id:
        raise ValueError("RUN_ID and TEMPLATE_ID environment variables must be set")

    execute_tests_and_store_results(run_id, template_id)
