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

from flask import Flask, jsonify, request, make_response
from flask_compress import Compress
from flask_httpauth import HTTPBasicAuth
from werkzeug.security import generate_password_hash, check_password_hash
from google.cloud import firestore
import os
from google.cloud import run_v2
from datetime import datetime
from util.settings import settings
from google.cloud import logging
import json

# setup logging
# Instantiates a client
logging_client = logging.Client()
# The name of the log to write to
log_name = "Litmus"
# Selects the log to write to
logger = logging_client.logger(log_name)
# Writes the log entry
logger.log_text("### Litmus starting ###")


def invoke_job(project_id, region, job_id, run_id, template_id):
    client = run_v2.JobsClient()
    job_name = client.job_path(project_id, region, job_id)

    override_spec = {
        "container_overrides": [
            {
                "env": [
                    {"name": "RUN_ID", "value": run_id},
                    {"name": "TEMPLATE_ID", "value": template_id},
                ]
            }
        ]
    }

    # Initialize the request
    job_name = f"projects/{project_id}/locations/{region}/jobs/{job_id}"
    request = run_v2.RunJobRequest(name=job_name, overrides=override_spec)

    response = client.run_job(request=request)
    return response
    # Handle response (e.g., check for errors, get execution details)


app = Flask(__name__, static_folder="ui/dist/", static_url_path="/")
auth = HTTPBasicAuth()
# Turn on compression
Compress(app)

if not settings.disable_auth:
    users = {settings.auth_user: generate_password_hash(settings.auth_pass)}

db = firestore.Client()


@auth.verify_password
def verify_password(username, password):
    if not settings.disable_auth:
        if username in users and check_password_hash(users.get(username), password):
            return username
    else:
        return username


@app.route("/version")
@auth.login_required
def version():
    # Version
    data = {"version": os.environ.get("VERSION", "0.0.0-alpha")}
    return make_response(jsonify(data), 200)


@app.route("/submit_run", methods=["POST"])
@auth.login_required
def submit_run():
    """Submits run with payload structure"""

    data = request.get_json()
    run_id = data.get("run_id")
    template_id = data.get("template_id")
    pre_request = data.get("pre_request")
    post_request = data.get("post_request")
    test_request = data.get("test_request")

    if not run_id or not template_id:
        return (
            jsonify({"error": "Missing 'run_id' or 'template_id' in request data"}),
            400,
        )

    if not test_request:
        return (
            jsonify({"error": "Missing 'test_request' in request data"}),
            400,
        )

    # Get test request template from Firestore
    template_ref = db.collection("test_templates").document(template_id)
    template_data = template_ref.get().to_dict()

    if not template_data:
        return jsonify({"error": f"Test template '{template_id}' not found"}), 404

    # Generate test requests with payload structure and URL
    tests = []
    for i, request_item in enumerate(template_data.get("template_data", [])):
        # Replace placeholders in test_request with values from request_item
        test = {
            "request": test_request,
        }
        if pre_request:
            if not isinstance(pre_request, dict):
                test["pre_request"] = json.loads(pre_request)
            else:
                test["pre_request"] = pre_request
        if post_request:
            if not isinstance(post_request, dict):
                test["post_request"] = json.loads(post_request)
            else:
                test["post_request"] = post_request

        json_string = json.dumps(test["request"])

        for key, value in request_item.items():

            # Replace the value
            json_string = json_string.replace(f"{{{key}}}", str(value))

            # Get the corresponding golden response from the template
            if key == "response":
                test["golden_response"] = value

        test["request"] = json.loads(json_string)

        if not isinstance(test["request"], dict):
            test["request"] = json.loads(test["request"])

        test["result"] = None

        tests.append(test)

    # Get current time for start timestamp
    start_time = datetime.utcnow()

    # Create a document for the test run (using the provided run_id)
    run_ref = db.collection("test_runs").document(run_id)
    run_ref.set(
        {
            "status": "Not Started",
            "progress": "0/0",
            "template_id": template_id,
            "start_time": start_time,
        }
    )

    # Store test cases in a subcollection (using the generated test requests)
    test_cases_collection = db.collection(f"test_cases_{run_id}")
    for i, request_data in enumerate(tests):
        test_case_ref = test_cases_collection.document(f"test_case_{i+1}")
        test_case_ref.set(request_data)

    project_id = settings.project_id
    region = settings.region
    job_id = "litmus-worker"
    invoke_job(project_id, region, job_id, run_id, template_id)

    return jsonify(
        {
            "message": f"Test run '{run_id}' submitted successfully using template '{template_id}'"
        }
    )


# Templates: Add
@app.route("/add_template", methods=["POST"])
@auth.login_required
def add_template():
    """Adds a new test template to Firestore with responses and pre/post requests."""
    data = request.get_json()
    template_id = data.get("template_id")
    template_data = data.get("template_data")
    test_pre_request = data.get("test_pre_request")
    test_post_request = data.get("test_post_request")
    test_request = data.get("test_request")

    if not template_id:
        return (
            jsonify(
                {"error": "Missing 'template_id' or 'template_data' in request data"}
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
            "test_request": test_request,
        }
    )
    return jsonify({"message": f"Template '{template_id}' added successfully"})


# Templates: Update
@app.route("/update_template", methods=["PUT"])
@auth.login_required
def update_template():
    """Updates an existing test template in Firestore, including responses and pre/post requests."""
    data = request.get_json()
    template_id = data.get("template_id")
    template_data = data.get("template_data")
    test_pre_request = data.get("test_pre_request")
    test_post_request = data.get("test_post_request")
    test_request = data.get("test_request")

    if not template_id:
        return jsonify({"error": "Missing 'template_id' in request data"}), 400

    template_ref = db.collection("test_templates").document(template_id)

    if not template_ref.get().exists:
        return jsonify({"error": f"Template with ID '{template_id}' not found"}), 404

    # Check if at least one field is being updated
    if not any(
        [
            template_data,
            test_pre_request,
            test_post_request,
            test_request,
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
    if test_request is not None:
        update_data["test_request"] = test_request

    template_ref.update(update_data)
    return jsonify({"message": f"Template '{template_id}' updated successfully"})


# Templates: Delete
@app.route("/delete_template/<template_id>", methods=["DELETE"])
@auth.login_required
def delete_template(template_id):
    """Deletes a test template from Firestore."""
    template_ref = db.collection("test_templates").document(template_id)
    if not template_ref.get().exists:
        return jsonify({"error": f"Template with ID '{template_id}' not found"}), 404
    template_ref.delete()
    return jsonify({"message": f"Template '{template_id}' deleted successfully"})


# Templates: List
@app.route("/templates", methods=["GET"])
@auth.login_required
def list_templates():
    """Lists all available test template IDs."""
    templates_ref = db.collection("test_templates")
    template_ids = [doc.id for doc in templates_ref.stream()]
    return jsonify({"template_ids": template_ids})


# Templates: Get
@app.route("/templates/<template_id>", methods=["GET"])
@auth.login_required
def get_template(template_id):
    """Retrieves a specific test template."""
    template_ref = db.collection("test_templates").document(template_id)
    template_data = template_ref.get().to_dict()
    if not template_data:
        return jsonify({"error": f"Template with ID '{template_id}' not found"}), 404
    return jsonify(template_data)


@app.route("/run_status/<run_id>", methods=["GET"])
@auth.login_required
def get_run_status(run_id):
    """Retrieves the status, progress, and detailed results (requests, responses) of a test run.
    Allows filtering of specific JSON paths to retrieve desired values.
    """
    run_ref = db.collection("test_runs").document(run_id)
    run_data = run_ref.get().to_dict()
    if not run_data:
        return jsonify({"error": f"Run with ID '{run_id}' not found"}), 404

    # Get test case details (including requests, responses, and golden answers)
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
        )  # Assuming "result" holds the actual response
        filtered_golden_response = filter_json(
            case_data.get("golden_response"), request.args.get("golden_response_filter")
        )

        test_cases.append(
            {
                "id": doc.id,  # Include the test case ID (e.g., "test_case_1")
                "request": filtered_request,
                "response": filtered_response,
                "golden_response": filtered_golden_response,
            }
        )

    return jsonify(
        {
            "status": run_data.get("status"),
            "progress": run_data.get("progress"),
            "testCases": test_cases,  # Return the detailed test case data
        }
    )


@app.route("/all_run_results/<template_id>", methods=["GET"])
@auth.login_required
def all_run_results(template_id):
    """Retrieves filtered responses for all runs of a specified template."""

    # Get all test runs for the given template
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
                    "data": filtered_response,  # Assuming "value" holds the actual response
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
        filter_pathx: A string representing the path to the desired values.
                     Uses dot notation (e.g., "key1.key2.key3")

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
                if key in current_data:
                    current_data = current_data[key]
                else:
                    # Key not found, move to the next path
                    break  # This is the key change

        # If we reached the end of the keys, add the current_data to new_data
        if current_data is not None:
            new_data[filter_path] = current_data

    return new_data


@app.route("/runs", methods=["GET"])
@auth.login_required
def list_runs():
    """Lists all test runs with their IDs and status, sorted by start time."""
    runs_ref = db.collection("test_runs")
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
            }
        )

    # Sort runs by start_time in descending order
    runs.sort(key=lambda run: run["start_time"], reverse=True)

    return jsonify({"runs": runs})


@app.route("/", defaults={"path": ""})
@app.route("/<path:path>")
@auth.login_required
def catch_all(path):
    return app.send_static_file("index.html")


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
