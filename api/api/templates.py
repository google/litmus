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

"""This module defines the API routes for test templates."""

from flask import Blueprint, jsonify, request
from flask_httpauth import HTTPBasicAuth
from google.cloud import firestore
from api.auth import auth
from util.settings import settings

bp = Blueprint("templates", __name__)
db = firestore.Client()


# Templates: Add
@bp.route("/add", methods=["POST"])
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
@bp.route("/update", methods=["PUT"])
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
@bp.route("/<template_id>", methods=["DELETE"])
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
@bp.route("/", methods=["GET"])
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
@bp.route("/<template_id>", methods=["GET"])
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
