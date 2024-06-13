# Litmus: HTTP Request and Response Testing Tool with User Interface & LLM Evaluation

Litmus is a comprehensive tool designed for testing and evaluating HTTP Requests and Responses. It combines a powerful API, a robust worker service, and a user-friendly web interface to streamline the testing process. 

## Features

- **Automated Test Execution:**  Submit test runs using pre-defined templates to evaluate responses against golden answers using AI.
- **Flexible Test Templates:** Define and manage test templates specifying the structure and parameters of your tests.
- **User-Friendly Web Interface:**  Interact with the Litmus platform through a visually appealing and intuitive web interface.
- **Detailed Results:**  View the status, progress, and detailed results of your test runs.
- **Advanced Filtering:** Filter responses from test runs based on specific JSON paths for in-depth analysis.
- **Performance Monitoring:** Track the performance of your responses and identify areas for improvement by using AI.
- **Cloud Integration:** Leverage the power of Google Cloud Platform (Firestore, Cloud Run) for efficient data storage and execution.
- **Quick Deployment:**  Use the provided deployment tool for a streamlined setup.

## Architecture

Litmus consists of three core components:

1. **API:**
   - Manages test templates, test runs, and user authentication.
   - Provides endpoints for submitting tests, retrieving results, and managing templates.
   - Uses Firestore for data storage.
2. **Worker Service:**
   - Executes test cases based on templates and provided test data.
   - Invokes the LLM and compares its responses against golden answers.
   - Stores test results in Firestore.
3. **User Interface:**
   - Allows users to interact with the Litmus platform.
   - Enables creating and managing test templates.
   - Presents test results in an organized and informative way.

## Getting Started

**1. Quick Deployment with the Deployment Tool:**

   - This is the easiest way to set up Litmus. 
   - Make sure you have the Google Cloud SDK installed and configured with the correct project.
   - Run the deployment script: 
     ```bash
     cd deploy
     ./deploy my-project us-central1 PASSWORD=test AI_DEFAULT_MODEL=gemini-1.5-flash
     ```
     - Replace placeholders with your actual values:
       - `<your-project-id>`: Your Google Cloud Project ID.
       - `<your-gcp-region>`: Your Google Cloud region.
       - `<your-password>`: A password for API authentication.
       - `<your-llm-model>`: The LLM model you want to use.

   - **The deployment script will:**
     - Create service accounts and grant necessary permissions.
     - Deploy the worker service and the API service to Cloud Run.

**2. Manual Setup:**

   - **If you prefer manual deployment:**
     - **Set up your Google Cloud project:**  Enable the required APIs (Firestore, Cloud Run).
     - **Deploy the worker service:** Build a Docker image for the worker service in the `worker` directory and deploy it to Cloud Run.
     - **Deploy the API service:** Build a Docker image for the API service in the `api` directory and deploy it to Cloud Run.
     - **Configure API settings:** Create a `api/util/settings.py` file with your Google Cloud project ID, region, and other settings.
     - **Deploy the UI:** Deploy the user interface code in the `api/ui` directory to a web server (e.g., Nginx, Apache).
     - **Connect the UI:** Configure the UI to connect to the deployed API service.

**3. Using Litmus:**

   - Access the web interface.
   - Create and manage test templates.
   - Submit test runs and analyze the results.


## Code Structure

- **api:** Contains the code for the API service.
- **ui:** Contains the user interface code.
- **worker:** Contains the code for the worker service.
- **deployment:** Contains deployment scripts to simplify the deployment process.


## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md) for details.

## License

Apache 2.0; see [`LICENSE`](LICENSE) for details.

## Disclaimer

This project is not an official Google project. It is not supported by
Google and Google specifically disclaims all warranties as to its quality,
merchantability, or fitness for a particular purpose.



