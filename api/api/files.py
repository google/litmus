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

"""This module defines the API routes for file management."""

from flask import Blueprint, jsonify, request, send_file
from werkzeug.utils import secure_filename
from google.cloud import storage
import os
from api.auth import auth
from util.settings import settings
from werkzeug.utils import secure_filename

bp = Blueprint("files", __name__)
# Initialize Storage client and get the bucket
storage_client = storage.Client()
files_bucket_name = os.environ.get("FILES_BUCKET")
files_bucket = storage_client.bucket(files_bucket_name)


@bp.route("/", methods=["GET"])
@auth.login_required
def list_files():
    """Lists all files in the files bucket, including their full GCS paths.

    Returns:
        JSON response containing a list of file details.
    """
    blobs = files_bucket.list_blobs()
    files = []
    for blob in blobs:
        files.append(
            {"name": blob.name, "gcs_path": f"gs://{files_bucket_name}/{blob.name}"}
        )
    return jsonify({"files": files}), 200


@bp.route("/<filename>", methods=["GET"])
@auth.login_required
def download_file(filename):
    """Downloads a file from the files bucket.

    Args:
        filename: The name of the file to download.

    Returns:
        The file content.
    """
    blob = files_bucket.blob(filename)
    if not blob.exists():
        return jsonify({"error": f"File '{filename}' not found"}), 404

    # Create a temporary file to store the downloaded content
    safe_filename = secure_filename(filename)
    temp_filename = os.path.join("/tmp", safe_filename)
    blob.download_to_filename(temp_filename)

    # Send the downloaded file to the client
    return send_file(temp_filename, as_attachment=True), 200


@bp.route("/<filename>", methods=["POST"])
@auth.login_required
def upload_file(filename):
    """Uploads a file to the files bucket.

    Args:
        filename: The name to store the uploaded file under.

    Returns:
        JSON response indicating success or failure.
    """
    if "file" not in request.files:
        return jsonify({"error": "No file part"}), 400

    file = request.files["file"]
    if file.filename == "":
        return jsonify({"error": "No selected file"}), 400

    blob = files_bucket.blob(filename)
    blob.upload_from_file(file)
    return jsonify({"message": f"File '{filename}' uploaded successfully"}), 201


@bp.route("/<filename>", methods=["DELETE"])
@auth.login_required
def delete_file(filename):
    """Deletes a file from the files bucket.

    Args:
        filename: The name of the file to delete.

    Returns:
        JSON response indicating success or failure.
    """
    blob = files_bucket.blob(filename)
    if not blob.exists():
        return jsonify({"error": f"File '{filename}' not found"}), 404

    blob.delete()
    return jsonify({"message": f"File '{filename}' deleted successfully"}), 200
