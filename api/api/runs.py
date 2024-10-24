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

"""This module defines the API routes for test runs and missions."""
import json
from datetime import datetime

from flask import Blueprint, jsonify, request, send_file
from google.cloud import firestore
from google.cloud import run_v2
from api.auth import auth
from util.settings import settings
from reportlab.lib.pagesizes import letter
from reportlab.platypus import SimpleDocTemplate, Paragraph, Spacer, PageBreak
from reportlab.lib.styles import getSampleStyleSheet
from io import BytesIO
from util.assess import ask_llm_for_summary

bp = Blueprint("runs", __name__)
db = firestore.Client()  # Initialize Firestore client


@bp.route("/submit", methods=["POST"])
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
        - evaluation_types (optional): Dictionary specifying evaluation methods.
                                        Example: {"ragas": True, "deepeval": ["answer_relevancy", "faithfulness"]}

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
    evaluation_types = data.get(
        "evaluation_types", {}
    )  # Get evaluation_types, default to empty dict

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
            if template_type == "Test Run":
                # Replace placeholders in test_request with values from request_item
                json_string = json_string.replace(f"{{{key}}}", str(value))
            elif template_type == "Test Mission" and key == "query":
                test["mission"] = value

            # Get the corresponding golden response from the template
            if key == "response":
                test["golden_response"] = value

        # Replace {auth_token} in the test request
        if auth_token:
            json_string = json_string.replace("{auth_token}", auth_token)

        test["request"] = json.loads(json_string)
        test["result"] = None
        test["flagged"] = False  # Initialize flagged to False
        test["rating"] = 0  # Initialize rating to 0
        test["comments"] = []  # Initialize comments to empty array
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
            "evaluation_types": evaluation_types,  # Store evaluation_types in the run data
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


@bp.route("/submit_simple", methods=["POST"])
@auth.login_required
def submit_simple():
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
        "template_llm_prompt": template_data.get("template_llm_prompt"),
        "template_input_field": template_data.get("template_input_field"),
        "template_output_field": template_data.get("template_output_field"),
        "template_type": template_data.get("template_type"),
        "evaluation_types": template_data.get("evaluation_types", {}),
        "auth_token": auth_token,
    }

    # Add mission_duration only if the template type is "Test Mission"
    if submit_data["template_type"] == "Test Mission":
        submit_data["mission_duration"] = template_data.get("mission_duration")

    # Call submit_run() with constructed data
    return submit_run(submit_data)


# Invoke Run
@bp.route("/invoke", methods=["POST"])
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
@bp.route("/<run_id>", methods=["DELETE"])
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


# Run Status
@bp.route("/status/<run_id>", methods=["GET"])
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

    # Apply filter for flagged test cases
    flagged_filter = request.args.get("flagged")
    if flagged_filter is not None:
        flagged_filter = flagged_filter.lower() == "true"
        test_cases_query = test_cases_query.where("flagged", "==", flagged_filter)

    # Apply filter for rating
    rating_filter = request.args.get("rating")
    if rating_filter is not None:
        try:
            rating_filter = int(rating_filter)
            test_cases_query = test_cases_query.where("rating", "==", rating_filter)
        except ValueError:
            return jsonify({"error": "Invalid rating filter value"}), 400

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

        # Get the comment count
        comment_count = len(case_data.get("comments", []))

        test_cases.append(
            {
                "id": doc.id,  # Include the test case ID (e.g., "test_case_1")
                "request": filtered_request,
                "response": filtered_response,
                "golden_response": filtered_golden_response,
                "tracing_id": case_data.get("tracing_id"),
                "flagged": case_data.get("flagged"),
                "rating": case_data.get("rating"),
                "comments": case_data.get("comments"),
                "comment_count": comment_count,  # Add comment count to the response
            }
        )

    return jsonify(
        {
            "status": run_data.get("status"),
            "progress": run_data.get("progress"),
            "template_id": run_data.get("template_id"),
            "template_input_field": run_data.get("template_input_field"),
            "template_output_field": run_data.get("template_output_field"),
            "template_type": run_data.get("template_type"),
            "testCases": test_cases,
        }
    )


@bp.route("/status_fields/<run_id>", methods=["GET"])
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


@bp.route("/all_results/<template_id>", methods=["GET"])
@auth.login_required
def all_results(template_id):
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


@bp.route("/", methods=["GET"])
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


# Flag a test case
@bp.route("/<run_id>/<case_id>/flag", methods=["PUT"])
@auth.login_required
def flag_test_case(run_id, case_id):
    """Toggles the flagged status of a test case.

    Args:
        run_id (str): The ID of the run containing the test case.
        case_id (str): The ID of the test case.

    Returns:
        A JSON response indicating success or failure.
    """
    case_ref = db.collection(f"test_cases_{run_id}").document(case_id)
    case_data = case_ref.get().to_dict()

    if not case_data:
        return (
            jsonify({"error": f"Test case '{case_id}' not found in run '{run_id}'"}),
            404,
        )

    current_flagged_status = case_data.get("flagged", False)
    case_ref.update({"flagged": not current_flagged_status})

    return (
        jsonify(
            {"message": f"Test case '{case_id}' flagged status updated successfully"}
        ),
        200,
    )


# Rate a test case
@bp.route("/<run_id>/<case_id>/rate", methods=["PUT"])
@auth.login_required
def rate_test_case(run_id, case_id):
    """Rates a test case.

    Args:
        run_id (str): The ID of the run containing the test case.
        case_id (str): The ID of the test case.

    Returns:
        A JSON response indicating success or failure.
    """

    data = request.get_json()
    rating = data.get("rating")

    if not rating:
        return jsonify({"error": "Missing 'rating' in request data"}), 400

    if rating not in range(1, 6):  # Rating should be between 1 and 5
        return jsonify({"error": "'rating' must be between 1 and 5"}), 400

    case_ref = db.collection(f"test_cases_{run_id}").document(case_id)
    case_data = case_ref.get().to_dict()

    if not case_data:
        return (
            jsonify({"error": f"Test case '{case_id}' not found in run '{run_id}'"}),
            404,
        )

    case_ref.update({"rating": rating})

    return jsonify({"message": f"Test case '{case_id}' rated successfully"}), 200


# Add or update a comment on a test case
@bp.route("/<run_id>/<case_id>/comment", methods=["POST"])
@auth.login_required
def comment_on_test_case(run_id, case_id):
    """Adds or updates a comment on a test case.

    Args:
        run_id (str): The ID of the run containing the test case.
        case_id (str): The ID of the test case.

    Returns:
        A JSON response indicating success or failure.
    """

    data = request.get_json()
    comment = data.get("comment")

    # Input validation
    if not comment:
        return jsonify({"error": "Missing 'comment' in request data"}), 400

    case_ref = db.collection(f"test_cases_{run_id}").document(case_id)
    case_data = case_ref.get().to_dict()

    if not case_data:
        return (
            jsonify({"error": f"Test case '{case_id}' not found in run '{run_id}'"}),
            404,
        )

    current_comments = case_data.get("comments", [])
    current_comments.append(comment)
    case_ref.update({"comments": current_comments})

    return (
        jsonify({"message": f"Comment added to test case '{case_id}' successfully"}),
        201,
    )


@bp.route("/report/<run_id>", methods=["GET"])
@auth.login_required
def generate_report(run_id):
    """Generates a PDF report for a given run ID."""

    try:
        run_ref = db.collection("test_runs").document(run_id)
        run_data = run_ref.get().to_dict()

        if not run_data:
            return jsonify({"error": f"Run with ID '{run_id}' not found"}), 404

        test_cases_ref = db.collection(f"test_cases_{run_id}")
        test_cases = []
        test_summaries = []

        for doc in test_cases_ref.stream():
            case_data = doc.to_dict()
            test_cases.append(case_data)

            test_summary = ask_llm_for_summary(
                question=case_data.get("request"),
                answer=case_data.get("result"),
                golden=case_data.get("golden_response"),
            )
            test_summaries.append(test_summary)

        # Create PDF
        buffer = BytesIO()
        doc = SimpleDocTemplate(buffer, pagesize=letter)
        styles = getSampleStyleSheet()
        story = []

        # Run Information
        story.append(Paragraph(f"Test Run Report: {run_id}", styles["Heading1"]))
        story.append(Spacer(1, 12))
        story.append(
            Paragraph(f"Template ID: {run_data.get('template_id')}", styles["Normal"])
        )
        story.append(
            Paragraph(
                f"Start Time: {run_data.get('start_time')}",
                styles["Normal"],
            )
        )
        story.append(
            Paragraph(
                f"End Time: {run_data.get('end_time')}",
                styles["Normal"],
            )
        )
        story.append(Spacer(1, 24))

        # Overall Summary
        overall_summary = ask_llm_for_summary(summaries=test_summaries)
        story.append(Paragraph(f"<b>{overall_summary}</b>", styles["Normal"]))
        story.append(Spacer(1, 24))

        overall_outliers = ask_llm_for_summary(outliers=test_summaries)
        story.append(Paragraph(f"<b>{overall_outliers}</b>", styles["Normal"]))
        story.append(Spacer(1, 24))

        story.append(PageBreak())

        for i, test_case in enumerate(test_cases):
            story.append(Paragraph(f"Test Case {i + 1}", styles["Heading1"]))
            story.append(Spacer(1, 12))

            story.append(Paragraph(f"<b>{test_summaries[i]}</b>", styles["Normal"]))
            story.append(Spacer(1, 12))

            request_json = json.dumps(test_case.get("request"), indent=2)
            story.append(
                Paragraph(
                    f"<i>Request: <pre>{request_json}</pre></i>", styles["Normal"]
                )
            )
            story.append(Spacer(1, 12))

            result = test_case.get("result")
            result_json = json.dumps(result, indent=2)
            story.append(
                Paragraph(f"<i>Result:<pre>{result_json}</pre></i>", styles["Normal"])
            )
            story.append(Spacer(1, 24))

            story.append(PageBreak())

        doc.build(story)
        buffer.seek(0)

        return send_file(
            buffer,
            as_attachment=True,
            download_name=f"report_{run_id}.pdf",
            mimetype="application/pdf",
        )

    except Exception as e:
        return jsonify({"error": f"Error generating report: {str(e)}"}), 500
