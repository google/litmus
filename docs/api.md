# Litmus API Reference

## Introduction

The Litmus API provides a programmatic interface to interact with the Litmus platform for testing and evaluating AI models, particularly Large Language Models (LLMs). It offers endpoints for submitting test runs, managing test templates, retrieving test results, accessing proxy data, and managing deployed proxy services.

## Authentication

The Litmus API uses basic authentication to protect its endpoints. To authenticate your requests, you need to include an `Authorization` header with your username and password encoded in Base64.

**Example:**

```
Authorization: Basic YWRtaW46YWRtaW4=
```

**Note:** The example above shows the Base64 encoded string for username "admin" and password "admin". Replace these with your actual credentials. You can disable authentication by setting `DISABLE_AUTH=True` as an environment variable for the API service.

## Endpoints

### 1. Submit Test Run

**Endpoint:** `/runs/submit`

**Method:** `POST`

**Description:** Submits a new test run using a pre-defined template.

**Request Body:**

```json
{
  "run_id": "string", // Unique identifier for the test run.
  "template_id": "string", // Identifier for the test template.
  "pre_request": { ... }, // Optional: Pre-request data (JSON object).
  "post_request": { ... }, // Optional: Post-request data (JSON object).
  "test_request": { ... } // Test request data with placeholders (JSON object).
}
```

**Response:**

```json
{
  "message": "string" // Success message indicating the run ID and template ID used.
}
```

**Status Codes:**

- `200 OK`: Test run submitted successfully.
- `400 Bad Request`: Missing required fields or invalid request data.
- `404 Not Found`: Test template not found.

### 2. Invoke Existing Test Run

**Endpoint:** `/runs/invoke`

**Method:** `POST`

**Description:** Re-invokes an existing test run.

**Request Body:**

```json
{
  "run_id": "string", // Unique identifier for the existing test run.
  "template_id": "string" // Identifier for the test template.
}
```

**Response:**

```json
{
  "message": "string" // Success message indicating the run ID and template ID used.
}
```

**Status Codes:**

- `200 OK`: Test run re-invoked successfully.
- `400 Bad Request`: Missing required fields or invalid request data.

### 3. Delete Test Run

**Endpoint:** `/runs/<run_id>`

**Method:** `DELETE`

**Description:** Deletes a test run and its associated test cases from Firestore.

**Path Parameters:**

- `run_id`: Unique identifier for the test run.

**Response:**

```json
{
  "message": "string" // Success message indicating the run ID that was deleted.
}
```

**Status Codes:**

- `200 OK`: Test run deleted successfully.
- `404 Not Found`: Run with the given ID not found.

### 4. Manage Test Templates

#### 4.1 Add Test Template

**Endpoint:** `/templates/add`

**Method:** `POST`

**Description:** Adds a new test template to Firestore.

**Request Body:**

```json
{
  "template_id": "string", // Unique identifier for the test template.
  "template_data": [ ... ], // Array of test data objects (JSON array).
  "test_pre_request": { ... }, // Optional: Pre-request data for the template (JSON object).
  "test_post_request": { ... }, // Optional: Post-request data for the template (JSON object).
  "test_request": { ... }, // Test request data with placeholders (JSON object).
  "template_llm_prompt": "string", // The LLM prompt associated with the test template.
  "template_input_field": "string", // The input field used in the test template.
  "template_output_field": "string" // The output field used in the test template.
}
```

**Response:**

```json
{
  "message": "string" // Success message indicating the template ID that was added.
}
```

**Status Codes:**

- `200 OK`: Template added successfully.
- `400 Bad Request`: Missing required fields or invalid request data.
- `409 Conflict`: Template with the given ID already exists.

#### 4.2 Update Test Template

**Endpoint:** `/templates/update`

**Method:** `PUT`

**Description:** Updates an existing test template in Firestore.

**Request Body:**

```json
{
  "template_id": "string", // Unique identifier for the test template.
  "template_data": [ ... ], // Optional: Updated array of test data objects (JSON array).
  "test_pre_request": { ... }, // Optional: Updated pre-request data (JSON object).
  "test_post_request": { ... }, // Optional: Updated post-request data (JSON object).
  "test_request": { ... }, // Optional: Updated test request data (JSON object).
  "template_llm_prompt": "string", // Optional: Updated LLM prompt.
  "template_input_field": "string", // Optional: Updated input field.
  "template_output_field": "string" // Optional: Updated output field.
}
```

**Response:**

```json
{
  "message": "string" // Success message indicating the template ID that was updated.
}
```

**Status Codes:**

- `200 OK`: Template updated successfully.
- `400 Bad Request`: Missing required fields or invalid request data.
- `404 Not Found`: Template with the given ID not found.

#### 4.3 Delete Test Template

**Endpoint:** `/templates/<template_id>`

**Method:** `DELETE`

**Description:** Deletes a test template from Firestore.

**Path Parameters:**

- `template_id`: Unique identifier for the test template.

**Response:**

```json
{
  "message": "string" // Success message indicating the template ID that was deleted.
}
```

**Status Codes:**

- `200 OK`: Template deleted successfully.
- `404 Not Found`: Template with the given ID not found.

#### 4.4 List Test Templates

**Endpoint:** `/templates`

**Method:** `GET`

**Description:** Retrieves a list of all available test template IDs.

**Response:**

```json
{
  "template_ids": [ "string", ... ] // Array of template IDs.
}
```

**Status Codes:**

- `200 OK`: List of templates retrieved successfully.

#### 4.5 Get Test Template

**Endpoint:** `/templates/<template_id>`

**Method:** `GET`

**Description:** Retrieves details of a specific test template.

**Path Parameters:**

- `template_id`: Unique identifier for the test template.

**Response:**

```json
{
  "template_data": [ ... ], // Array of test data objects (JSON array).
  "test_pre_request": { ... }, // Optional: Pre-request data for the template (JSON object).
  "test_post_request": { ... }, // Optional: Post-request data for the template (JSON object).
  "test_request": { ... }, // Test request data with placeholders (JSON object).
  "template_llm_prompt": "string", // The LLM prompt associated with the test template.
  "template_input_field": "string", // The input field used in the test template.
  "template_output_field": "string" // The output field used in the test template.
}
```

**Status Codes:**

- `200 OK`: Template details retrieved successfully.
- `404 Not Found`: Template with the given ID not found.

### 5. Retrieve Test Run Details

#### 5.1 Get Run Status

**Endpoint:** `/runs/status/<run_id>`

**Method:** `GET`

**Description:** Retrieves the status and detailed results of a test run, optionally filtering the data based on JSON paths.

**Path Parameters:**

- `run_id`: Unique identifier for the test run.

**Query Parameters:**

- `request_filter`: Comma-separated string of JSON paths to filter the request data (e.g., `body.param1,headers.Authorization`).
- `response_filter`: Comma-separated string of JSON paths to filter the response data (e.g., `result.status,assessment.similarity`).
- `golden_response_filter`: Comma-separated string of JSON paths to filter the golden response data.

**Response:**

```json
{
  "status": "string", // Status of the test run (e.g., "Completed", "Running").
  "progress": "string", // Progress of the test run (e.g., "10/10").
  "template_id": "string", // ID of the template used for the run.
  "template_input_field": "string", // The input field used in the test template.
  "template_output_field": "string", // The output field used in the test template.
  "testCases": [
    {
      "id": "string", // Test case ID (e.g., "test_case_1").
      "request": { ... }, // Filtered request data (JSON object).
      "response": { ... }, // Filtered response data (JSON object).
      "golden_response": { ... }, // Filtered golden response data (JSON object).
      "tracing_id": "string" // Unique ID for tracing the request in proxy logs.
    },
    ...
  ]
}
```

**Status Codes:**

- `200 OK`: Run status and details retrieved successfully.
- `404 Not Found`: Run with the given ID not found.

#### 5.2 Get Run Status Fields

**Endpoint:** `/runs/status_fields/<run_id>`

**Method:** `GET`

**Description:** Retrieves specific fields from the status of a test run, including date, template ID, and input/output fields.

**Path Parameters:**

- `run_id`: Unique identifier for the test run.

**Response:**

```json
{
  "run_date": "string", // Date of the run (ISO 8601 format).
  "template_id": "string", // Template ID used for the run.
  "template_input_field": "string", // Input field used in the template.
  "template_output_field": "string" // Output field used in the template.
}
```

**Status Codes:**

- `200 OK`: Run status fields retrieved successfully.
- `404 Not Found`: Run with the given ID not found.

#### 5.3 Get Results for All Runs of a Template

**Endpoint:** `/runs/all_results/<template_id>`

**Method:** `GET`

**Description:** Retrieves filtered responses for all runs of a specified template, grouped by the value of the request filter.

**Path Parameters:**

- `template_id`: The ID of the test template.

**Query Parameters:**

- `request_filter`: JSON path to filter the request data (e.g., `body.param1`).
- `response_filter`: Comma-separated string of JSON paths to filter the response data (e.g., `result.status,assessment.similarity`).

**Response:**

```json
{
  "request_value1": [ // Grouped by unique values from the `request_filter`
    {
      "start_time": "string", // Start time of the run (ISO 8601 format).
      "end_time": "string", // End time of the run (ISO 8601 format), if available.
      "run_id": "string", // Run ID.
      "data": { ... } // Filtered response data (JSON object).
    },
    ...
  ],
  "request_value2": [ ... ], // Another group of responses for a different request value
  ...
}
```

**Status Codes:**

- `200 OK`: Results for all runs retrieved successfully.

### 6. Retrieve Proxy Data

#### 6.1 Get Proxy Data

**Endpoint:** `/proxy/data`

**Method:** `GET`

**Description:** Retrieves proxy log data from BigQuery for a specific date, optionally filtered by Litmus context and flattened.

**Query Parameters:**

- `date`: Date of the log entries (format: YYYY-MM-DD).
- `context` (optional): Litmus context to filter the results (e.g., `litmus-context-your-context-id`).
- `flatten` (optional): Whether to flatten the JSON structure of the results (default: `False`).

**Response:**

- **If `flatten` is `False` (default):** Returns an array of JSON objects, each representing a log entry.
- **If `flatten` is `True`:** Returns an array of flattened JSON objects, where nested keys are concatenated with underscores (e.g., `requestHeaders_user_agent`).

**Status Codes:**

- `200 OK`: Proxy data retrieved successfully.
- `400 Bad Request`: Missing `date` parameter.
- `500 Internal Server Error`: Error querying BigQuery.

#### 6.2 Get Aggregated Proxy Data

**Endpoint:** `/proxy/agg`

**Method:** `GET`

**Description:** Retrieves aggregated proxy log data from BigQuery for a specific date, optionally filtered by Litmus context.

**Query Parameters:**

- `date`: Date of the log entries (format: YYYY-MM-DD).
- `context` (optional): Litmus context to filter the results (e.g., `litmus-context-your-context-id`).

**Response:**

```json
[
  {
    "litmuscontext": "string", // Litmus context.
    "requestheaders_x_goog_request_params": "string", // Request parameters from Google's API.
    "total_token_count": "integer", // Total token count for the requests in the group.
    "prompt_token_count": "integer", // Prompt token count for the requests in the group.
    "candidates_token_count": "integer", // Candidates token count for the requests in the group.
    "average_latency": "float" // Average latency for the requests in the group (in milliseconds).
  },
  ...
]
```

**Status Codes:**

- `200 OK`: Aggregated proxy data retrieved successfully.
- `400 Bad Request`: Missing `date` parameter.
- `500 Internal Server Error`: Error querying BigQuery.

### 7. List Proxy Services

**Endpoint:** `/proxy/list_services`

**Method:** `GET`

**Description:** Retrieves a list of deployed Litmus proxy Cloud Run services.

**Response:**

```json
{
  "proxy_services": [
    {
      "name": "string", // Name of the proxy service.
      "project_id": "string", // GCP Project ID.
      "region": "string", // GCP Region.
      "uri": "string", // URI of the proxy service.
      "created": "string", // Creation timestamp (ISO 8601 format).
      "updated": "string" // Last update timestamp (ISO 8601 format).
    },
    ...
  ]
}
```

**Status Codes:**

- `200 OK`: List of proxy services retrieved successfully.
- `500 Internal Server Error`: Error listing Cloud Run services.

### 8. Version

**Endpoint:** `/version`

**Method:** `GET`

**Description:** Returns the version of the Litmus API.

**Response:**

```json
{
  "version": "string" // Version number of the API.
}
```

**Status Codes:**

- `200 OK`: Version retrieved successfully.

## Error Handling

The Litmus API uses standard HTTP status codes to indicate the success or failure of a request. In case of an error, the response body will typically include a JSON object with an `error` field providing a description of the error.

**Example Error Response:**

```json
{
  "error": "Missing 'run_id' in request data"
}
```
