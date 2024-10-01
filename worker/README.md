## Litmus Worker

This repository contains the code for the Litmus worker, a Cloud Run job responsible for executing test runs based on predefined templates and storing the results in Firestore. The worker supports both single-turn interactions ("Test Run" templates) and multi-turn conversations ("Test Mission" templates) where an LLM dynamically generates requests.

### Features

- **Test Execution:** Executes test requests based on the defined template data and user-provided test cases, handling different HTTP methods (POST, GET, PUT, DELETE).
- **Result Storage:** Stores test results, including responses, assessment details, and status codes, in a Firestore database.
- **Progress Updates:** Tracks the progress of the test run and updates the run status in Firestore in real time.
- **Multiple Evaluation Methods:** Supports various methods for evaluating LLM responses:
  - **Custom LLM Assessment:** Uses a separate LLM to compare the actual response with a golden response based on a customizable prompt.
  - **Ragas Evaluation:** Applies a suite of Ragas metrics, such as answer relevancy, context recall, context precision, harmfulness, and answer similarity.
  - **DeepEval Evaluation:** Utilizes DeepEval's LLM-based metrics to assess aspects like faithfulness, contextual relevance, hallucination, bias, and toxicity.
- **Pre/Post Request Execution:** Allows for optional pre-request and post-request hooks to be defined within the template. These hooks are executed before and after each test case, enabling setup and teardown actions for the testing environment.
- **File Reference Handling:** Supports the use of file references (e.g., "[FILE: my_file.txt]") within test cases and templates. The worker automatically retrieves the content of referenced files from a configured Google Cloud Storage bucket.
- **Tracing:** Assigns unique tracing IDs to each request, facilitating correlation with proxy logs for comprehensive analysis and debugging.

### Prerequisites

- **Google Cloud Project:** An active GCP project with billing enabled.
- **Firestore Database:** A Firestore database within your GCP project for storing test templates and run data.
- **BigQuery Dataset:** A BigQuery dataset (named `litmus_analytics`) for storing and analyzing proxy logs.
- **Google Cloud Storage Bucket:** A GCS bucket for storing files referenced in test cases and templates.
- **Service Accounts:** Service accounts for the API service and the worker job. The worker's service account must have the following permissions:
  - **Firestore:** Read and write access to the Firestore database.
  - **Cloud Run Invoker:** Permission to invoke the Litmus API Cloud Run service.
  - **BigQuery Data Viewer:** Read access to the BigQuery dataset for retrieving proxy logs.
  - **Storage Object Viewer:** Read access to the GCS bucket containing referenced files.
- **Vertex AI API Enabled:** The Vertex AI API must be enabled in your GCP project to utilize Vertex AI models for Ragas and DeepEval evaluations.
- **Environment Variables:** The worker Cloud Run job should have the following environment variables set:
  - `GCP_PROJECT`: Your GCP project ID.
  - `GCP_REGION`: The GCP region where your Litmus resources are deployed.
  - `FILES_BUCKET`: The name of your GCS bucket containing referenced files.
  - `FILES_PREFIX` (optional): A prefix for file paths within your GCS bucket (defaults to no prefix).

### Deployment

1. **Build Docker Image:**

   - Navigate to the `worker` directory.
   - Build the Docker image:
     ```bash
     docker build -t gcr.io/<your-project-id>/litmus-worker:latest .
     ```

2. **Push Image to Container Registry:**

   - Push the built image to your container registry (e.g., Google Container Registry):
     ```bash
     docker push gcr.io/<your-project-id>/litmus-worker:latest
     ```

3. **Deploy as Cloud Run Job:**
   - Deploy the worker as a Cloud Run _job_ (not a service), ensuring you set the required environment variables:
     ```bash
     gcloud run jobs deploy litmus-worker \
         --image=gcr.io/<your-project-id>/litmus-worker:latest \
         --region=<your-gcp-region> \
         --set-env-vars=GCP_PROJECT=<your-project-id>,GCP_REGION=<your-gcp-region>,FILES_BUCKET=<your-files-bucket> \
         --service-account=<your-worker-service-account>
     ```

### Triggering Test Runs

The Litmus API is responsible for triggering the worker job whenever a new test run is submitted. It passes the `RUN_ID` and `TEMPLATE_ID` as environment variables to the worker, providing the necessary information to execute the tests.

### Code Structure

- **`main.py`:** The main worker script that orchestrates test execution, result storage, progress updates, and calls to other modules for LLM evaluation and file handling.
- **`util/assess.py`:** Contains functions for interacting with LLMs for assessments, including:
  - `ask_llm_against_golden`: Compares an actual response with a golden response using a custom prompt.
  - `ask_llm_for_action`: Generates the next request for a "Test Mission" based on the mission description and conversation history.
  - `is_mission_done`: Checks if the mission is complete using the LLM.
  - `evaluate_mission`: Evaluates the overall success of a "Test Mission" using the LLM.
- **`util/ragas_eval.py`:** Handles Ragas evaluation using the specified metrics.
- **`util/deepeval_eval.py`:** Handles DeepEval evaluation using the selected DeepEval metrics.
- **`util/settings.py`:** Contains configuration settings for the worker, including GCP project ID and location, and the AI model to use for evaluation.
- **`util/docsnsnips.py`:** Provides utility functions for working with documents and citations, including extracting article IDs from URLs, renumbering citations, and handling citations in LLM responses.

### Configuration and Extension

- You can customize the AI models used for evaluation, default evaluation methods, and other settings by modifying the `util/settings.py` file.
- To add new evaluation methods or extend existing ones (Ragas, DeepEval), you'll need to modify the corresponding modules within the `util` directory. These modifications require rebuilding and redeploying the worker Docker image.

### Note

- This README provides a concise overview of the Litmus worker. For detailed information on specific functions and configurations, refer to the comprehensive code comments within the worker script and its associated modules.
