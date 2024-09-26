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

"""This module defines the API routes for proxy data."""

from flask import Blueprint, jsonify, request
from flask_httpauth import HTTPBasicAuth
from google.cloud import bigquery
from google.cloud import run_v2
from datetime import datetime
from api.auth import auth
from util.settings import settings

bp = Blueprint("proxy", __name__)
bq_client = bigquery.Client()


@bp.route("/data", methods=["GET"])
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


@bp.route("/litmus_data", methods=["GET"])
@auth.login_required
def litmus_data():
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
        FROM `{settings.project_id}.litmus_analytics.litmus_core_log_{date}`
        ORDER BY jsonPayload.timestamp ASC
        LIMIT 1000
    """

    if context:
        query = f"""
            SELECT jsonPayload
            FROM `{settings.project_id}.litmus_analytics.litmus_core_log_{date}`
            WHERE jsonPayload.requestheaders.x_litmus_request = "{context}"
            ORDER BY jsonPayload.timestamp ASC
            LIMIT 1000
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


@bp.route("/agg", methods=["GET"])
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


@bp.route("/list_services", methods=["GET"])
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
