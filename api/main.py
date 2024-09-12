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

"""This module defines the Flask application for Litmus, a tool for testing and monitoring LLMs."""
import json
import os
from datetime import datetime

from flask import Flask, jsonify, request, make_response
from flask_compress import Compress
from flask_httpauth import HTTPBasicAuth
from google.cloud import firestore, logging, bigquery
from google.cloud import run_v2
from werkzeug.security import generate_password_hash, check_password_hash

from util.settings import settings

# Setup logging
logging_client = logging.Client()
log_name = "Litmus"
logger = logging_client.logger(log_name)
logger.log_text("### Litmus starting ###")

# Flask app initialization
app = Flask(__name__, static_folder="ui/dist/", static_url_path="/")
auth = HTTPBasicAuth()
# Turn on compression
Compress(app)

# Authentication setup
if not settings.disable_auth:
    users = {settings.auth_user: generate_password_hash(settings.auth_pass)}

# Initialize Firestore and BigQuery clients
db = firestore.Client()
bq_client = bigquery.Client()


# Authentication verification callback function
@auth.verify_password
def verify_password(username, password):
    """Verifies user password for basic authentication.

    Args:
        username: Username provided in the request.
        password: Password provided in the request.

    Returns:
        Username if authentication is successful, otherwise None.
    """
    if not settings.disable_auth:
        return (
            username
            if username in users and check_password_hash(users.get(username), password)
            else None
        )
    return username  # Disable authentication


@app.route("/version")
@auth.login_required
def version():
    """Returns the version of the application.

    Returns:
        JSON response containing the version number.
    """
    return make_response(
        jsonify({"version": os.environ.get("VERSION", "0.0.0-alpha")}), 200
    )


@app.route("/submit_run", methods=["POST"])
@auth.login_required
def submit_run(data=None):
    """Submits a new test run or test mission.

    Can be called with data from a request or directly with a data dictionary.

    Args:
        data (dict, optional): A dictionary containing run information.
                             If None, data is fetched from request.get_json().

    Expects a JSON payload with:
        - run_id: Unique identifier for the test run/mission.
        - template_id: Identifier for the test template.
        - pre_request (optional): JSON object representing a pre-request to be executed.
        - post_request (optional): JSON object representing a post-request to be executed.
        - test_request: JSON object representing the test request.
        - template_llm_prompt: The LLM prompt associated with the test template.
        - template_input_field: The input field used in the test template.
        - template_output_field: The output field used in the test template.
        - template_type (optional): The type of template. Can be "Test Run" or "Test Mission".
                                     Defaults to "Test Run" if not provided.
        - mission_duration (optional): Number of interaction loops for a "Test Mission". Required if
                                        template_type is "Test Mission".

    Returns:
        JSON response indicating success or failure.
    """

    if data is None:
        data = request.get_json()
    run_id = data.get("run_id")
    template_id = data.get("template_id")
    pre_request = data.get("pre_request")
    post_request = data.get("post_request")
    test_request = data.get("test_request")
    auth_token = data.get("auth_token")

    # Input validation
    if not run_id or not template_id:
        return (
            jsonify({"error": "Missing 'run_id' or 'template_id' in request data"}),
            400,
        )

    if not test_request:
        return jsonify({"error": "Missing 'test_request' in request data"}), 400

    # Retrieve test template from Firestore
    template_ref = db.collection("test_templates").document(template_id)
    template_data = template_ref.get().to_dict()
    template_type = template_data.get("template_type")

    if not template_data:
        return jsonify({"error": f"Test template '{template_id}' not found"}), 404

    # Generate test requests with payload structure and URL
    tests = []
    for i, request_item in enumerate(template_data.get("template_data", [])):
        test = {}
        if test_request:
            test["request"] = (
                json.loads(test_request)
                if not isinstance(test_request, dict)
                else test_request
            )
        if pre_request:
            test["pre_request"] = (
                json.loads(pre_request)
                if not isinstance(pre_request, dict)
                else pre_request
            )
        if post_request:
            test["post_request"] = (
                json.loads(post_request)
                if not isinstance(post_request, dict)
                else post_request
            )

        json_string = json.dumps(test["request"])

        for key, value in request_item.items():
            # Replace placeholders in test_request with values from request_item
            json_string = json_string.replace(f"{{{key}}}", str(value))

            # Get the corresponding golden response from the template
            if key == "response":
                test["golden_response"] = value

        # Replace {auth_token} in the test request
        if auth_token:
            json_string = json_string.replace("{auth_token}", auth_token)

        test["request"] = json.loads(json_string)
        test["result"] = None
        tests.append(test)

    # Get current time for start timestamp
    start_time = datetime.utcnow()

    # Create a document for the test run/mission
    run_ref = db.collection("test_runs").document(run_id)
    run_ref.set(
        {
            "status": "Not Started",
            "progress": "0/{}".format(len(tests)),
            "template_id": template_id,
            "start_time": start_time,
            "template_input_field": template_data.get("template_input_field"),
            "template_output_field": template_data.get("template_output_field"),
            "template_llm_prompt": template_data.get("template_llm_prompt"),
            "template_type": template_type,
            "mission_duration": template_data.get("mission_duration"),
        }
    )

    # Store test cases in a subcollection
    test_cases_collection = db.collection(f"test_cases_{run_id}")
    for i, request_data in enumerate(tests):
        test_case_ref = test_cases_collection.document(f"test_case_{i+1}")
        test_case_ref.set(request_data)

    # Invoke the Cloud Run job to execute the test run/mission
    invoke_job(
        settings.project_id,
        settings.region,
        "litmus-worker",
        run_id,
        template_id,
        template_type,
        template_data.get("mission_duration"),
    )

    return jsonify(
        {
            "message": f"{template_type} '{run_id}' submitted successfully using template '{template_id}'"
        }
    )


@app.route("/submit_run_simple", methods=["POST"])
@auth.login_required
def submit_run_simple():
    """Submits a new test run or mission using default values from the template.

    Expects a JSON payload with:
        - run_id: Unique identifier for the test run/mission.
        - template_id: Identifier for the test template.

    Returns:
        JSON response indicating success or failure.
    """

    data = request.get_json()
    run_id = data.get("run_id")
    template_id = data.get("template_id")
    auth_token = data.get("auth_token")

    # Input validation
    if not run_id or not template_id:
        return (
            jsonify({"error": "Missing 'run_id' or 'template_id' in request data"}),
            400,
        )

    # Retrieve test template from Firestore
    template_ref = db.collection("test_templates").document(template_id)
    template_data = template_ref.get().to_dict()

    if not template_data:
        return jsonify({"error": f"Test template '{template_id}' not found"}), 404

    # Construct data for submit_run() using template defaults
    submit_data = {
        "run_id": run_id,
        "template_id": template_id,
        "test_request": template_data.get("test_request"),
        "pre_request": template_data.get("test_pre_request"),
        "post_request": template_data.get("test_post_request"),
        "auth_token": auth_token,
    }

    # Call submit_run() with constructed data
    return submit_run(submit_data)


# Invoke Run
@app.route("/invoke_run", methods=["POST"])
@auth.login_required
def invoke_run():
    """Re-invokes an existing test run or mission.

    Expects a JSON payload with:
        - run_id: Unique identifier for the test run/mission.
        - template_id: Identifier for the test template.

    Returns:
        JSON response indicating success or failure.
    """

    data = request.get_json()
    run_id = data.get("run_id")
    template_id = data.get("template_id")

    # Input validation
    if not run_id or not template_id:
        return (
            jsonify({"error": "Missing 'run_id' or 'template_id' in request data"}),
            400,
        )

    # Retrieve the template type and mission duration from the existing run data
    run_ref = db.collection("test_runs").document(run_id)
    run_data = run_ref.get().to_dict()
    if not run_data:
        return jsonify({"error": f"Run with ID '{run_id}' not found"}), 404

    template_type = run_data.get("template_type", "Test Run")
    mission_duration = run_data.get("mission_duration")

    # Update run status to "Not Started"
    run_ref.update({"status": "Not Started", "progress": "0/0"})

    # Invoke the Cloud Run job to execute the test run/mission
    invoke_job(
        settings.project_id,
        settings.region,
        "litmus-worker",
        run_id,
        template_id,
        template_type,
        mission_duration,
    )

    return jsonify(
        {
            "message": f"{template_type} '{run_id}' submitted successfully using template '{template_id}'"
        }
    )


# Delete a run
@app.route("/delete_run/<run_id>", methods=["DELETE"])
@auth.login_required
def delete_run(run_id):
    """Deletes a test run or mission from Firestore.

    Args:
        run_id: Unique identifier for the test run/mission.

    Returns:
        JSON response indicating success or failure.
    """

    # Delete test cases from the subcollection
    test_cases_collection = db.collection(f"test_cases_{run_id}")
    for doc in test_cases_collection.stream():
        doc.reference.delete()

    # Delete the run document itself
    run_ref = db.collection("test_runs").document(run_id)
    if run_ref.get().exists:
        run_ref.delete()
        return jsonify({"message": f"Run '{run_id}' deleted successfully"})
    return jsonify({"error": f"Run with ID '{run_id}' not found"}), 404


# Templates: Add
@app.route("/add_template", methods=["POST"])
@auth.login_required
def add_template():
    """Adds a new test template to Firestore.

    Expects a JSON payload with:
        - template_id: Unique identifier for the test template.
        - template_data: Array of test data objects.
        - test_pre_request (optional): JSON object representing a pre-request to be executed.
        - test_post_request (optional): JSON object representing a post-request to be executed.
        - test_request: JSON object representing the test request.
        - template_llm_prompt: The LLM prompt associated with the test template.
        - template_input_field: The input field used in the test template.
        - template_output_field: The output field used in the test template.
        - template_type: The type of template. Can be "Test Run" or "Test Mission".
        - mission_duration (optional): Number of interaction loops for a "Test Mission". Required if
                                        template_type is "Test Mission".

    Returns:
        JSON response indicating success or failure.
    """
    data = request.get_json()
    template_id = data.get("template_id")
    template_data = data.get("template_data")
    test_pre_request = data.get("test_pre_request")
    test_post_request = data.get("test_post_request")
    test_request = data.get("test_request")
    template_llm_prompt = data.get("template_llm_prompt")
    template_input_field = data.get("template_input_field")
    template_output_field = data.get("template_output_field")
    template_type = data.get("template_type")  # "Test Run" or "Test Mission"
    mission_duration = (
        data.get("mission_duration") if template_type == "Test Mission" else None
    )

    # Input validation
    if not template_id:
        return (
            jsonify(
                {"error": "Missing 'template_id' or 'template_data' in request data"}
            ),
            400,
        )

    if not template_type or template_type not in ["Test Run", "Test Mission"]:
        return (
            jsonify(
                {
                    "error": "'template_type' is required and must be either 'Test Run' or 'Test Mission'"
                }
            ),
            400,
        )

    if template_type == "Test Mission" and not isinstance(mission_duration, int):
        return (
            jsonify(
                {
                    "error": "'mission_duration' is required and must be an integer for 'Test Mission' type"
                }
            ),
            400,
        )

    # Check if a template with the same ID already exists
    template_ref = db.collection("test_templates").document(template_id)
    if template_ref.get().exists:
        return (
            jsonify({"error": f"Template with ID '{template_id}' already exists"}),
            409,
        )

    # Create the new template document
    template_ref.set(
        {
            "template_data": template_data,
            "test_pre_request": test_pre_request if test_pre_request else None,
            "test_post_request": test_post_request if test_post_request else None,
            "template_llm_prompt": template_llm_prompt if template_llm_prompt else None,
            "test_request": test_request,
            "template_input_field": template_input_field,
            "template_output_field": template_output_field,
            "template_type": template_type,  # Store the template type
            "mission_duration": mission_duration,  # Store mission duration, if applicable
        }
    )
    return jsonify({"message": f"Template '{template_id}' added successfully"})


# Templates: Update
@app.route("/update_template", methods=["PUT"])
@auth.login_required
def update_template():
    """Updates an existing test template in Firestore.

    Expects a JSON payload with:
        - template_id: Unique identifier for the test template.
        - template_data (optional): Array of test data objects.
        - test_pre_request (optional): JSON object representing a pre-request to be executed.
        - test_post_request (optional): JSON object representing a post-request to be executed.
        - test_request (optional): JSON object representing the test request.
        - template_llm_prompt (optional): The LLM prompt associated with the test template.
        - template_input_field (optional): The input field used in the test template.
        - template_output_field (optional): The output field used in the test template.
        - template_type (optional): The type of template. Can be "Test Run" or "Test Mission".
        - mission_duration (optional): Number of interaction loops for a "Test Mission". Required if
                                        template_type is updated to "Test Mission" and not previously set.

    Returns:
        JSON response indicating success or failure.
    """
    data = request.get_json()
    template_id = data.get("template_id")
    template_data = data.get("template_data")
    test_pre_request = data.get("test_pre_request")
    test_post_request = data.get("test_post_request")
    test_request = data.get("test_request")
    template_llm_prompt = data.get("template_llm_prompt")
    template_input_field = data.get("template_input_field")
    template_output_field = data.get("template_output_field")
    template_type = data.get("template_type")
    mission_duration = data.get("mission_duration")

    # Input validation
    if not template_id:
        return jsonify({"error": "Missing 'template_id' in request data"}), 400

    template_ref = db.collection("test_templates").document(template_id)

    if not template_ref.get().exists:
        return jsonify({"error": f"Template with ID '{template_id}' not found"}), 404

    # Check if template_type is being updated to "Test Mission" without mission_duration
    if (
        template_type == "Test Mission"
        and not mission_duration
        and not template_ref.get().to_dict().get("mission_duration")
    ):
        return (
            jsonify(
                {
                    "error": "'mission_duration' is required when updating template_type to 'Test Mission'"
                }
            ),
            400,
        )

    # Check if at least one field is being updated
    if not any(
        [
            template_data,
            test_pre_request,
            test_post_request,
            template_llm_prompt,
            test_request,
            template_input_field,
            template_output_field,
            template_type,
            mission_duration,
        ]
    ):
        return jsonify({"error": "No fields provided for update"}), 400

    update_data = {}
    if template_data is not None:
        update_data["template_data"] = template_data
    if test_pre_request is not None:
        update_data["test_pre_request"] = test_pre_request
    if test_post_request is not None:
        update_data["test_post_request"] = test_post_request
    if template_llm_prompt is not None:
        update_data["template_llm_prompt"] = template_llm_prompt
    if test_request is not None:
        update_data["test_request"] = test_request
    if template_input_field is not None:
        update_data["template_input_field"] = template_input_field
    if template_output_field is not None:
        update_data["template_output_field"] = template_output_field
    if template_type is not None:
        update_data["template_type"] = template_type
    if mission_duration is not None:
        update_data["mission_duration"] = mission_duration

    template_ref.update(update_data)
    return jsonify({"message": f"Template '{template_id}' updated successfully"})


# Templates: Delete
@app.route("/delete_template/<template_id>", methods=["DELETE"])
@auth.login_required
def delete_template(template_id):
    """Deletes a test template from Firestore.

    Args:
        template_id: Unique identifier for the test template.

    Returns:
        JSON response indicating success or failure.
    """
    template_ref = db.collection("test_templates").document(template_id)
    if not template_ref.get().exists:
        return jsonify({"error": f"Template with ID '{template_id}' not found"}), 404
    template_ref.delete()
    return jsonify({"message": f"Template '{template_id}' deleted successfully"})


# Templates: List
@app.route("/templates", methods=["GET"])
@auth.login_required
def list_templates():
    """Retrieves a list of all available test template IDs, optionally filtered by type.

    Query parameters:
        - type (optional): The type of templates to retrieve ("Test Run" or "Test Mission").
                           If not provided, returns all templates.

    Returns:
        JSON response containing an array of template IDs and their types.
    """
    template_type_filter = request.args.get("type")

    templates_ref = db.collection("test_templates")

    if template_type_filter:
        # Apply template type filter
        templates_ref = templates_ref.where("template_type", "==", template_type_filter)

    templates = []
    for doc in templates_ref.stream():
        template_data = doc.to_dict()
        templates.append(
            {
                "template_id": doc.id,
                "template_type": template_data.get("template_type"),
            }
        )
    return jsonify({"templates": templates})


# Templates: Get
@app.route("/templates/<template_id>", methods=["GET"])
@auth.login_required
def get_template(template_id):
    """Retrieves details of a specific test template.

    Args:
        template_id: Unique identifier for the test template.

    Returns:
        JSON response containing template details.
    """
    template_ref = db.collection("test_templates").document(template_id)
    template_data = template_ref.get().to_dict()
    if not template_data:
        return jsonify({"error": f"Template with ID '{template_id}' not found"}), 404
    return jsonify(template_data)


@app.route("/run_status/<run_id>", methods=["GET"])
@auth.login_required
def get_run_status(run_id):
    """Retrieves the status and detailed results of a test run/mission.

    Args:
        run_id: Unique identifier for the test run/mission.

    Returns:
        JSON response containing run status, progress, and test case details.
    """
    run_ref = db.collection("test_runs").document(run_id)
    run_data = run_ref.get().to_dict()
    if not run_data:
        return jsonify({"error": f"Run with ID '{run_id}' not found"}), 404

    # Get test case details
    test_cases_query = db.collection(f"test_cases_{run_id}")
    test_cases = []
    for doc in test_cases_query.stream():
        case_data = doc.to_dict()
        # Filter requests, responses, and golden responses based on query parameters
        filtered_request = filter_json(
            case_data.get("request"), request.args.get("request_filter")
        )
        filtered_response = filter_json(
            case_data.get("result"), request.args.get("response_filter")
        )
        filtered_golden_response = filter_json(
            case_data.get("golden_response"), request.args.get("golden_response_filter")
        )

        test_cases.append(
            {
                "id": doc.id,  # Include the test case ID (e.g., "test_case_1")
                "request": filtered_request,
                "response": filtered_response,
                "golden_response": filtered_golden_response,
                "tracing_id": case_data.get("tracing_id"),
            }
        )

    return jsonify(
        {
            "status": run_data.get("status"),
            "progress": run_data.get("progress"),
            "template_id": run_data.get("template_id"),
            "template_input_field": run_data.get("template_input_field"),
            "template_output_field": run_data.get("template_output_field"),
            "template_type": run_data.get("template_type"),  # Include template type
            "testCases": test_cases,  # Return the detailed test case data
        }
    )


@app.route("/run_status_fields/<run_id>", methods=["GET"])
@auth.login_required
def get_run_status_fields(run_id):
    """Retrieves specific fields from the status of a test run/mission.

    Args:
        run_id: Unique identifier for the test run/mission.

    Returns:
        JSON response containing run date, template ID, input/output fields, and template type.
    """
    run_ref = db.collection("test_runs").document(run_id)
    run_data = run_ref.get().to_dict()
    if not run_data:
        return jsonify({"error": f"Run with ID '{run_id}' not found"}), 404

    return jsonify(
        {
            "run_date": run_data.get("start_time"),
            "template_id": run_data.get("template_id"),
            "template_input_field": run_data.get("template_input_field"),
            "template_output_field": run_data.get("template_output_field"),
            "template_type": run_data.get("template_type"),  # Include template type
        }
    )


@app.route("/all_run_results/<template_id>", methods=["GET"])
@auth.login_required
def all_run_results(template_id):
    """Retrieves filtered responses for all runs/missions of a specified template.

    Args:
        template_id: The ID of the test template.

    Returns:
        JSON response containing a dictionary of results, keyed by request filter value.
        Each value is a list of responses sorted by start time, including run and time information.
    """

    # Get all test runs/missions for the given template
    runs_ref = db.collection("test_runs")
    runs = []
    for doc in runs_ref.stream():
        run_data = doc.to_dict()
        if run_data.get("template_id") == template_id:
            runs.append(
                {
                    "run_id": doc.id,
                    "start_time": run_data.get("start_time"),
                    "end_time": run_data.get(
                        "end_time"
                    ),  # Include end_time if available
                    "progress": run_data.get("progress"),
                    "template_type": run_data.get("template_type"),
                }
            )

    # Initialize results dictionary
    results = {}

    # Iterate through each run and gather filtered responses
    for run in runs:
        run_id = run["run_id"]
        test_cases_query = db.collection(f"test_cases_{run_id}")
        for doc in test_cases_query.stream():
            case_data = doc.to_dict()
            # Filter request and response based on query parameters
            filtered_request = filter_json(
                case_data.get("request"), request.args.get("request_filter")
            )
            filtered_response = filter_json(
                case_data.get("result"), request.args.get("response_filter")
            )

            # Add results to the dictionary with the start_time and run_id included
            if request.args.get("request_filter") in filtered_request:
                request_key = str(
                    filtered_request[request.args.get("request_filter")]
                )  # Use string representation for key
            if request_key not in results:
                results[request_key] = []
            results[request_key].append(
                {
                    "start_time": run["start_time"],
                    "end_time": run["end_time"],
                    "run_id": run_id,
                    "template_type": run["template_type"],  # Include template type here
                    "data": filtered_response,
                }
            )

    # Sort results within each request key by start_time
    for request_key in results:
        results[request_key].sort(key=lambda item: item["start_time"])

    # Return the results in the desired format
    return jsonify(results)


def filter_json(data, filter_pathx):
    """Filters a JSON structure based on the given path.

    Args:
        data: The JSON data to filter.
        filter_pathx: A comma-separated string representing the paths to the desired values.
                    Uses dot notation (e.g., "key1.key2.key3") and supports
                    array indexing (e.g., "key1.key2[0].key3").

    Returns:
        A dictionary containing the filtered values.
    """

    if not filter_pathx:
        return data  # No filtering needed

    filter_paths = filter_pathx.split(",")
    new_data = {}

    for filter_path in filter_paths:
        keys = filter_path.split(".")
        current_data = data
        for key in keys:
            if current_data is not None:
                # Check for array indexing
                if "[" in key and "]" in key:
                    key, index = key.split("[", 1)[0], int(key.split("[", 1)[1][:-1])
                    current_data = (
                        current_data[key][index]
                        if 0 <= index < len(current_data) and key in current_data
                        else None
                    )
                elif key in current_data:
                    current_data = current_data[key]
                else:
                    # Key not found, move to the next path
                    current_data = None
                    break

        # If we reached the end of the keys, add the current_data to new_data
        if current_data is not None:
            new_data[filter_path] = current_data

    return new_data


@app.route("/runs", methods=["GET"])
@auth.login_required
def list_runs():
    """Lists all test runs/missions with their details, sorted by start time.

    Query parameters:
        - type (optional): The type of runs/missions to retrieve ("Test Run" or "Test Mission").
                           If not provided, returns all runs/missions.

    Returns:
        JSON response containing an array of run/mission details.
    """
    runs_ref = db.collection("test_runs")

    # Apply filtering if type is provided in query parameters
    template_type_filter = request.args.get("type")
    if template_type_filter:
        runs_ref = runs_ref.where("template_type", "==", template_type_filter)

    runs = []
    for doc in runs_ref.stream():
        run_data = doc.to_dict()
        runs.append(
            {
                "run_id": doc.id,
                "status": run_data.get("status"),
                "start_time": run_data.get("start_time"),
                "end_time": run_data.get("end_time"),
                "progress": run_data.get("progress"),
                "template_id": run_data.get("template_id"),
                "template_type": run_data.get("template_type"),  # Include template type
            }
        )

    # Sort runs by start_time in descending order
    runs.sort(key=lambda run: run["start_time"], reverse=True)

    return jsonify({"runs": runs})


@app.route("/proxy_data", methods=["GET"])
@auth.login_required
def proxy_data():
    """Retrieves proxy log data from BigQuery.

    Expects query parameters:
        - date: Date of the log data (format: YYYY-MM-DD).
        - context (optional): Litmus context to filter the results.
        - flatten (optional): Whether to flatten the JSON structure of the results (default: False).

    Returns:
        JSON response containing the proxy log data.
    """
    date = request.args.get("date")
    context = request.args.get("context")

    # Get the 'flatten' flag from query parameters
    flatten_results = request.args.get(
        "flatten", default=False, type=lambda v: v.lower() == "true"
    )

    # Input validation
    if not date:
        return jsonify({"error": 'Missing "date" or "context" parameter'}), 400

    query = f"""
        SELECT jsonPayload
        FROM `{settings.project_id}.litmus_analytics.litmus_proxy_log_{date}`
        ORDER BY jsonPayload.timestamp ASC
        LIMIT 100
    """

    if context:
        query = f"""
            SELECT jsonPayload
            FROM `{settings.project_id}.litmus_analytics.litmus_proxy_log_{date}`
            WHERE jsonPayload.litmuscontext = "{context}"
            ORDER BY jsonPayload.timestamp ASC
            LIMIT 100
        """

    try:
        query_job = bq_client.query(query)
        results = list(query_job.result())

        # Flatten the results if requested
        if flatten_results:
            processed_results = []
            for row in results:
                flattened_row = flatten_json(row.jsonPayload)
                processed_results.append(flattened_row)
        else:
            # Return data as is if not flattening
            processed_results = [row.jsonPayload for row in results]

        return jsonify(processed_results)
    except Exception as e:
        return jsonify({"error": f"Error querying BigQuery: {str(e)}"}), 500


@app.route("/proxy_agg", methods=["GET"])
@auth.login_required
def proxy_agg():
    """Retrieves aggregated proxy log data from BigQuery.

    Expects query parameters:
        - date: Date of the log data (format: YYYY-MM-DD).
        - context (optional): Litmus context to filter the results.

    Returns:
        JSON response containing the aggregated proxy log data.
    """
    date = request.args.get("date")
    context = request.args.get("context")

    # Input validation
    if not date:
        return jsonify({"error": 'Missing "date" parameter'}), 400

    query = f"""
        SELECT
            jsonPayload.litmuscontext,
            jsonPayload.requestheaders.x_goog_request_params,
            sum(jsonPayload.responsebody.usagemetadata.totaltokencount) AS total_token_count,
            sum(jsonPayload.responsebody.usagemetadata.prompttokencount) AS prompt_token_count,
            sum(jsonPayload.responsebody.usagemetadata.candidatestokencount) AS candidates_token_count,
            avg(jsonPayload.latency) AS average_latency
        FROM
            `{settings.project_id}.litmus_analytics.litmus_proxy_log_{date}`
        GROUP BY 1,2;
    """

    if context:
        query = f"""
            SELECT
                jsonPayload.litmuscontext,
                jsonPayload.requestheaders.x_goog_request_params,
                sum(jsonPayload.responsebody.usagemetadata.totaltokencount) AS total_token_count,
                sum(jsonPayload.responsebody.usagemetadata.prompttokencount) AS prompt_token_count,
                sum(jsonPayload.responsebody.usagemetadata.candidatestokencount) AS candidates_token_count,
                avg(jsonPayload.latency) AS average_latency
            FROM
                `{settings.project_id}.litmus_analytics.litmus_proxy_log_{date}`
            WHERE jsonPayload.litmuscontext = "{context}"
            GROUP BY 1,2;
        """

    try:
        query_job = bq_client.query(query)
        results = list(query_job.result())

        # Convert BigQuery Rows to dictionaries before JSON serialization
        formatted_results = []
        for row in results:
            formatted_row = dict(row.items())
            formatted_results.append(formatted_row)

        return jsonify(formatted_results)

    except Exception as e:
        return jsonify({"error": f"Error querying BigQuery: {str(e)}"}), 500


@app.route("/", defaults={"path": ""})
@app.route("/<path:path>")
@auth.login_required
def catch_all(path):
    """Catches all undefined routes and serves the index.html file."""
    return app.send_static_file("index.html")


@app.route("/list_proxy_services", methods=["GET"])
@auth.login_required
def list_proxy_services():
    """Retrieves a list of deployed Litmus proxy Cloud Run services.

    Returns:
        JSON response containing an array of proxy service details.
    """
    project_id = settings.project_id
    region = settings.region

    client = run_v2.ServicesClient()

    request = run_v2.ListServicesRequest(
        parent=f"projects/{project_id}/locations/{region}"
    )
    try:
        page_result = client.list_services(request=request)
        services = list(page_result)

        proxy_services = []
        for service in services:
            if "aiplatform-litmus" in service.name:
                # Extract just the service name from the full path
                name = service.name.split("/")[-1]
                uri = service.uri
                created = datetime.fromtimestamp(service.create_time.timestamp())
                updated = datetime.fromtimestamp(service.update_time.timestamp())
                proxy_services.append(
                    {
                        "name": name,
                        "project_id": project_id,
                        "region": region,
                        "uri": uri,
                        "created": created,
                        "updated": updated,
                    }
                )

        return jsonify(proxy_services)

    except Exception as e:
        return jsonify({"error": f"Error listing Cloud Run services: {str(e)}"}), 500


def flatten_json(data, parent_key="", sep="_"):
    """
    Recursively flattens a nested JSON structure into a single-level dictionary.

    Args:
        data: The JSON data to flatten.
        parent_key: The key prefix for nested objects (used in recursive calls).
        sep: The separator to use between nested keys.

    Returns:
        A flattened dictionary.
    """
    items = []
    if isinstance(data, dict):
        for k, v in data.items():
            new_key = f"{parent_key}{sep}{k}" if parent_key else k
            items.extend(flatten_json(v, new_key, sep=sep).items())
    elif isinstance(data, list):
        for i, v in enumerate(data):
            new_key = f"{parent_key}{sep}{i}" if parent_key else str(i)
            items.extend(flatten_json(v, new_key, sep=sep).items())
    else:
        items.append((parent_key, data))
    return dict(items)


def invoke_job(
    project_id, region, job_id, run_id, template_id, template_type, mission_duration
):
    """Invokes a Cloud Run job with specified parameters.

    Args:
        project_id: Google Cloud project ID.
        region: Google Cloud region.
        job_id: Cloud Run job ID.
        run_id: Unique identifier for the test run/mission.
        template_id: Identifier for the test template.
        template_type: The type of template ("Test Run" or "Test Mission").
        mission_duration: Number of interaction loops for a "Test Mission" (can be None).
    """
    client = run_v2.JobsClient()
    job_name = client.job_path(project_id, region, job_id)

    # Include template_type and mission_duration in environment variables
    env_vars = [
        {"name": "RUN_ID", "value": run_id},
        {"name": "TEMPLATE_ID", "value": template_id},
        {"name": "TEMPLATE_TYPE", "value": template_type},
    ]
    if mission_duration is not None:
        env_vars.append(
            {"name": "MISSION_DURATION", "value": str(mission_duration)}
        )  # Ensure mission_duration is a string

    override_spec = {"container_overrides": [{"env": env_vars}]}

    # Initialize the request
    job_name = f"projects/{project_id}/locations/{region}/jobs/{job_id}"
    request = run_v2.RunJobRequest(name=job_name, overrides=override_spec)

    client.run_job(request=request)


if __name__ == "__main__":
    app.run(debug=False, host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
