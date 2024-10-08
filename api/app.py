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

"""This module defines the Flask application for Litmus."""

import os

from flask import Flask, jsonify
from flask_compress import Compress
from google.cloud import logging
from api import runs, templates, proxy, files, auth


# Setup logging
logging_client = logging.Client()
log_name = "Litmus"
logger = logging_client.logger(log_name)
logger.log_text("### Litmus starting ###")

# Flask app initialization
app = Flask(__name__, static_folder="ui/dist/", static_url_path="/")
app.url_map.strict_slashes = False
# Turn on compression
Compress(app)

# Register blueprints
app.register_blueprint(runs.bp, url_prefix="/runs")
app.register_blueprint(templates.bp, url_prefix="/templates")
app.register_blueprint(proxy.bp, url_prefix="/proxy")
app.register_blueprint(files.bp, url_prefix="/files")


# Version
@app.route("/version")
@auth.login_required
def version():
    """Returns the version of the application."""
    return jsonify({"version": os.environ.get("VERSION", "0.0.0-alpha")}), 200


# Serving Static Files
@app.route("/", defaults={"path": ""})
@app.route("/<path:path>")
@auth.login_required
def catch_all(path):
    """Catches all undefined routes and serves the index.html file."""
    return app.send_static_file("index.html")


if __name__ == "__main__":
    app.run(debug=True, host="0.0.0.0", port=int(os.environ.get("PORT", 8080)))
