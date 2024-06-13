## Litmus Worker

This repository contains the code for the Litmus worker, a Cloud Run job that executes test runs and stores the results.

**Features:**

- **Test Execution:** Executes test requests based on the defined template data and user-provided test cases.
- **Result Storage:** Stores test results, including responses and assessments, in a Firestore database.
- **Progress Updates:** Tracks the progress of the test run and updates the run status in Firestore.
- **LLM-Based Assessment:** Uses a large language model (LLM) to assess the similarity between actual responses and golden responses.
- **Pre/Post Request Execution:** Supports optional pre and post requests that can be executed before and after the main test request.

**Prerequisites:**

- **Google Cloud Project:** You need a Google Cloud project to deploy and run the Litmus worker.
- **Firestore Database:**  Create a Firestore database in your Google Cloud project.
- **Service Accounts:** Create a service account for the worker with permission to access Firestore and invoke the Litmus API.
- **Cloud Run Job:** Deploy the worker as a Cloud Run job.
- **Docker Image:** Build and push the Docker image for the worker to a container registry (e.g., Google Container Registry).
- **Google Cloud SDK:** Ensure that you have the Google Cloud SDK installed and configured with the correct project.
- **LLM Model:** The `ask_llm_against_golden` function in `util/assess.py` requires an LLM model. Ensure that you have access to and are configured to use a suitable model.

**Getting Started:**

1. **Set Up Google Cloud Project:**
   - Enable the required APIs (Firestore, Cloud Run, etc.).
   - Create a service account for the worker.
   - Grant necessary permissions to the worker service account (Firestore access, API invocation permissions).

2. **Deploy the Worker:**
   - Build the Docker image for the worker:
     ```bash
     docker build -t gcr.io/<your-project-id>/litmus-worker:latest .
     ```
   - Push the image to your container registry:
     ```bash
     docker push gcr.io/<your-project-id>/litmus-worker:latest
     ```
   - Deploy the worker to Cloud Run:
     ```bash
     gcloud run deploy litmus-worker --image gcr.io/<your-project-id>/litmus-worker:latest --region <your-gcp-region>
     ```

3. **Trigger Test Runs:**
   - The Litmus API will invoke the worker job using Cloud Run, passing the `RUN_ID` and `TEMPLATE_ID` as environment variables.

**Code Structure:**

- **`execute_request`:**  Executes an HTTP request and returns the response.
- **`execute_tests_and_store_results`:**  Handles the execution of tests, stores the results in Firestore, and updates run progress.
- **`ask_llm_against_golden` (in `util/assess.py`):** Compares an actual response with a golden response using an LLM.

**Note:**  This README provides a basic overview of the Litmus worker. For detailed documentation about the code and functionality, refer to the code comments within the worker script and the `util/assess.py` file.
