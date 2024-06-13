## Litmus API

This repository contains the code for the Litmus API, a service that manages and orchestrates test runs for AI models. It provides endpoints to:

- **Submit test runs:** Create new test runs using pre-defined templates, including user-provided test data and optional pre/post requests.
- **Manage test templates:** Create, update, delete, and list test templates in Firestore, defining the structure and parameters of test runs.
- **Retrieve test run status and results:** Get the status, progress, and detailed results (requests, responses) of a test run.
- **Filter results:** Filter responses from test runs based on specific JSON paths.
- **Retrieve results for all runs of a template:**  Gather filtered responses from all runs of a specified template, sorted by start time.

**Features:**

- **Firestore Integration:** Stores test templates and test run data in a Firestore database.
- **Cloud Run Job Invocation:** Uses Google Cloud Run to execute test runs using a dedicated worker job.
- **Authorization:** Implements basic authentication to protect API endpoints.
- **CORS Support:** Enables cross-origin requests from web applications.
- **Compression:**  Uses GZIP compression to optimize API responses.
- **Filter JSON Paths:**  Provides functionality to filter responses based on specific JSON paths for easier analysis.

**Prerequisites:**

- **Google Cloud Project:** You need a Google Cloud project to deploy and run the Litmus API.
- **Firestore Database:** Create a Firestore database in your Google Cloud project.
- **Service Accounts:** Create service accounts for both the API and the worker, and grant the worker permission to invoke the API.
- **Cloud Run Job:** Deploy a Cloud Run job that handles the execution of test cases.
- **Docker Image:** Build and push the Docker image for the API to a container registry (e.g., Google Container Registry).
- **Google Cloud SDK:** Ensure that you have the Google Cloud SDK installed and configured with the correct project.

**Getting Started:**

1. **Set Up Your Google Cloud Project:**
   - Enable the required APIs (Firestore, Cloud Run, etc.).
   - Create service accounts and grant necessary permissions.
   - Deploy the worker job using a Docker image.

2. **Configure Settings:**
   - Create a `util/settings.py` file with the following settings:
     ```python
     project_id = "your-google-cloud-project-id"
     region = "your-gcp-region"
     disable_auth = False # Set to True to disable authentication
     auth_pass = "your-password" # Set this to a password you want to use for basic auth
     ```

3. **Build and Deploy the API:**
   - Build the Docker image for the API:
     ```bash
     docker build -t gcr.io/<your-project-id>/litmus-api:latest .
     ```
   - Push the image to your container registry:
     ```bash
     docker push gcr.io/<your-project-id>/litmus-api:latest
     ```
   - Deploy the API to Cloud Run:
     ```bash
     gcloud run deploy litmus-api --image gcr.io/<your-project-id>/litmus-api:latest --region <your-gcp-region> --allow-unauthenticated
     ```

4. **Use the API:**
   - Once deployed, you can access the API endpoints using your preferred HTTP client or tools like Postman.
   - Consult the API documentation within the code for detailed information about each endpoint and its usage.

**API Endpoints:**

**1. Submit Test Run**

   - **Endpoint:** `/submit_run`
   - **Method:** POST
   - **Request Body:**
     ```json
     {
       "run_id": "your-run-id",
       "template_id": "your-template-id",
       "pre_request": { ... }, // Optional: Pre-request data
       "post_request": { ... }, // Optional: Post-request data
       "test_request": { ... } //  Test request data with placeholders
     }
     ```
   - **Response:**
     ```json
     {
       "message": "Test run 'your-run-id' submitted successfully using template 'your-template-id'"
     }
     ```

**2. Manage Test Templates**

   - **Add Template:**
     - **Endpoint:** `/add_template`
     - **Method:** POST
     - **Request Body:**
       ```json
       {
         "template_id": "your-template-id",
         "template_data": [ ... ], // Array of request/response pairs with placeholders
         "test_pre_request": { ... }, // Optional: Pre-request for the template
         "test_post_request": { ... }, // Optional: Post-request for the template
         "test_request": { ... } // Test request with placeholders
       }
       ```
     - **Response:**
       ```json
       {
         "message": "Template 'your-template-id' added successfully"
       }
       ```

   - **Update Template:**
     - **Endpoint:** `/update_template`
     - **Method:** PUT
     - **Request Body:**
       ```json
       {
         "template_id": "your-template-id",
         "template_data": [ ... ], // Updated template data (optional)
         "test_pre_request": { ... }, // Updated pre-request (optional)
         "test_post_request": { ... }, // Updated post-request (optional)
         "test_request": { ... } // Updated test request (optional)
       }
       ```
     - **Response:**
       ```json
       {
         "message": "Template 'your-template-id' updated successfully"
       }
       ```

   - **Delete Template:**
     - **Endpoint:** `/delete_template/<template_id>`
     - **Method:** DELETE
     - **Response:**
       ```json
       {
         "message": "Template 'your-template-id' deleted successfully"
       }
       ```

   - **List Templates:**
     - **Endpoint:** `/templates`
     - **Method:** GET
     - **Response:**
       ```json
       {
         "template_ids": [ "template1", "template2", ... ]
       }
       ```

   - **Get Template:**
     - **Endpoint:** `/templates/<template_id>`
     - **Method:** GET
     - **Response:**
       ```json
       {
         "template_data": [ ... ], // Array of request/response pairs
         "test_pre_request": { ... }, // Pre-request for the template (optional)
         "test_post_request": { ... }, // Post-request for the template (optional)
         "test_request": { ... } // Test request for the template 
       }
       ```

**3. Retrieve Test Run Status and Results**

   - **Get Run Status:**
     - **Endpoint:** `/run_status/<run_id>`
     - **Method:** GET
     - **Query Parameters:**
       - `request_filter`:  Filter the request based on a JSON path (e.g., `body.param1`).
       - `response_filter`: Filter the response based on a JSON path (e.g., `result.status`).
       - `golden_response_filter`: Filter the golden response based on a JSON path.
     - **Response:**
       ```json
       {
         "status": "Completed",
         "progress": "10/10",
         "testCases": [
           {
             "id": "test_case_1", // ID of the test case
             "request": { ... }, // Filtered request data
             "response": { ... }, // Filtered response data
             "golden_response": { ... } // Filtered golden response data 
           },
           ...
         ]
       }
       ```

   - **Get Results for All Runs of a Template:**
     - **Endpoint:** `/all_run_results/<template_id>`
     - **Method:** GET
     - **Query Parameters:**
       - `request_filter`:  Filter the request based on a JSON path (e.g., `body.param1`).
       - `response_filter`: Filter the response based on a JSON path (e.g., `result.status`).
     - **Response:**
       ```json
       {
         "request_key1": [ 
           {
             "start_time": "2024-06-05T14:20:39.000Z",
             "end_time": "2024-06-05T14:20:42.000Z", 
             "run_id": "run-123",
             "data": { ... } // Filtered response data
           },
           {
             "start_time": "2024-06-05T14:25:39.000Z",
             "end_time": "2024-06-05T14:25:42.000Z", 
             "run_id": "run-456",
             "data": { ... } // Filtered response data
           },
           ...
         ],
         "request_key2": [ ... ], 
         ...
       }
       ```

**4. List All Test Runs**

   - **Endpoint:** `/runs`
   - **Method:** GET
   - **Response:**
     ```json
     {
       "runs": [
         {
           "run_id": "your-run-id",
           "status": "Completed",
           "start_time": "2024-06-05T14:20:39.000Z",
           "end_time": "2024-06-05T14:20:42.000Z",
           "progress": "10/10",
           "template_id": "your-template-id"
         },
         ...
       ]
     }
     ```

**5. Version**

   - **Endpoint:** `/version`
   - **Method:** GET
   - **Response:**
     ```json
     {
       "version": "0.0.0-alpha"
     }
     ```

**Authentication:**

- The API uses basic authentication to protect its endpoints.
- You can disable authentication by setting `disable_auth` to `True` in `util/settings.py`.
- If authentication is enabled, you need to send basic authentication credentials (username and password) in the request headers.

**Note:** 
- This documentation outlines the main API endpoints and their basic functionality. 
- Refer to the API code for more detailed documentation about each endpoint and its usage.

