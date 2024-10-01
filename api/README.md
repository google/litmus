## Litmus API

This repository contains the code for the Litmus API, a service that manages and orchestrates test runs and evaluations for AI models, particularly Large Language Models (LLMs). It provides a programmatic interface for various functionalities, enabling you to integrate Litmus into your AI testing workflows.

## Key Features

- **Test Run & Mission Management (`api/runs.py`):**
  - Create new test runs based on pre-defined templates, including user-provided data and optional pre/post requests.
  - Initiate "Test Missions," which involve multi-turn interactions guided by an LLM, enabling more realistic conversational or task-oriented evaluations.
  - Retrieve the status, progress, and detailed results (requests, responses) of test runs.
  - Restart or delete existing test runs.
- **Test Template Management (`api/templates.py`):**
  - Define reusable test templates, specifying the structure and parameters for test runs.
  - Create, update, delete, and list test templates, allowing you to organize and manage your test configurations.
  - Include pre/post requests within templates for setup and teardown actions.
  - Define LLM prompts for AI-driven assessments of responses.
  - Specify input and output fields within the request and response payloads for targeted evaluation.
- **Flexible Evaluation Options:**

  - Choose from multiple LLM evaluation methods within your test templates:
    - **Custom LLM Evaluation (`worker/util/assess.py`):** Use customizable prompts to guide an LLM in assessing the quality of responses compared to expected outputs.
    - **Ragas Evaluation (`worker/util/ragas_eval.py`):** Leverage Ragas metrics, including answer relevancy, context recall, context precision, harmfulness, and answer similarity.
    - **DeepEval Evaluation (`worker/util/deepeval_eval.py`):** Utilize DeepEval's LLM-based metrics, such as answer relevancy, faithfulness, contextual precision, contextual recall, hallucination, bias, and toxicity.

- **Proxy Data Access and Analysis (`api/proxy.py`):**

  - Retrieve detailed and aggregated proxy log data from BigQuery, providing insights into LLM usage patterns, performance, and potential issues.
  - Access proxy data filtered by date, context, or specific fields.

- **Proxy Service Management (`api/proxy.py`):**

  - List and manage deployed Litmus proxy services that capture LLM interactions.
  - Retrieve details about each proxy service, including its name, URI, creation time, and last update time.

- **File Management (`api/files.py`):**

  - Upload, download, list, and delete files associated with your test cases and templates.
  - Reference these files directly in your test data, making your JSON payloads more concise and manageable.

- **Test Case Interaction (`api/runs.py`):**

  - Flag and rate individual test cases to highlight potential issues or noteworthy results.
  - Add comments to test cases for collaboration and documentation purposes.

- **Robust and Secure:**
  - Implements basic authentication (`api/auth.py`) to protect API endpoints.
  - Uses GZIP compression for efficient data transfer.

## Technologies

- **Flask:** A lightweight and flexible web framework for Python.
- **Firestore:** A NoSQL document database for storing test templates and run data.
- **Cloud Run:** A serverless platform for running the API and worker services.
- **BigQuery:** A data warehouse for storing and analyzing proxy logs.
- **Vertex AI:** A platform for accessing and utilizing Google's powerful LLMs for evaluation purposes.
- **Google Cloud Storage:** A cloud storage service for managing files referenced in test cases and templates.

## Prerequisites

- **Google Cloud Project:** An active GCP project with billing enabled.
- **Enabled APIs:** Ensure the following APIs are enabled in your GCP project:
  - Cloud Run API
  - Firestore API
  - BigQuery API
  - Cloud Resource Manager API
  - Vertex AI API
  - Secret Manager API
  - Cloud Storage API
- **Service Accounts:** Create service accounts for the API and worker, granting the worker permission to invoke the API and access necessary resources (Firestore, BigQuery, Cloud Storage).
- **Cloud Run Job:** Deploy the worker service as a Cloud Run _job_ (see the worker service documentation for instructions).
- **Docker Image:** Build and push a Docker image for the API to a container registry (e.g., Google Container Registry).
- **Settings Configuration:** Create a `api/util/settings.py` file to configure essential settings, including:
  - GCP project ID
  - GCP region
  - AI model settings (location, default model, validation model)
  - Authentication settings (enable/disable, username, password)
- **UI Deployment (optional):** Deploy the Litmus UI to a web server to provide a graphical interface for interacting with the API (see the UI documentation for instructions).

## Deployment

1. **Build Docker Image:**
   - Navigate to the `api` directory.
   - Build the Docker image:
     ```bash
     docker build -t gcr.io/<your-project-id>/litmus-api:latest .
     ```
2. **Push Docker Image:**
   - Push the image to your container registry:
     ```bash
     docker push gcr.io/<your-project-id>/litmus-api:latest
     ```
3. **Deploy to Cloud Run:**
   - Deploy the API as a Cloud Run _service_:
     ```bash
     gcloud run deploy litmus-api \
         --image=gcr.io/<your-project-id>/litmus-api:latest \
         --region=<your-gcp-region> \
         --allow-unauthenticated \
         --service-account=<your-api-service-account>
     ```

## Usage

Once deployed, you can interact with the Litmus API using your preferred HTTP client or tools like Postman. Refer to the [API Documentation](https://google.github.io/litmus/api) for details about each endpoint, expected payloads, and responses.

## Authentication

- The API uses basic authentication to protect its endpoints.
- To authenticate, include an `Authorization` header in your requests with the username and password encoded in Base64.
- You can disable authentication by setting `DISABLE_AUTH=True` as an environment variable for the API service. **Caution:** Disabling authentication is not recommended for production environments.

## Error Handling

The API utilizes standard HTTP status codes to indicate the success or failure of requests. In case of an error, the response body will typically include a JSON object with an `error` field providing a description of the error.

## Contributions

We welcome contributions to the Litmus API! If you'd like to contribute, please refer to the [Contribution Guide](https://google.github.io/litmus/contribution) file for guidelines and instructions.
