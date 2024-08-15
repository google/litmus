# Litmus: HTTP Request and Response Testing Tool with User Interface & LLM Evaluation

![DEV](https://github.com/google/litmus/actions/workflows/dev_deploy.yml/badge.svg)
![UAT](https://github.com/google/litmus/actions/workflows/uat_deploy.yml/badge.svg)
![PROD](https://github.com/google/litmus/actions/workflows/prod_deploy.yml/badge.svg)

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

**1. Quick Deployment with the Litmus CLI:**

   - This is the easiest way to set up Litmus. 
   - Make sure you have the Google Cloud SDK installed and configured with the correct project.
   
   - Install the Litmus CLI:
      - **Linux**:
         - install:```curl https://storage.googleapis.com/litmus-cloud/install/linux.sh | sudo sh```
         - binary: [https://storage.googleapis.com/litmus-cloud/prod/linux/litmus](https://storage.googleapis.com/litmus-cloud/prod/linux/litmus)
         - sha256: [https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256)
      - **OSX**:
         - install:```curl https://storage.googleapis.com/litmus-cloud/install/osx.sh | sudo sh```
         - binary: [https://storage.googleapis.com/litmus-cloud/prod/osx/litmus](https://storage.googleapis.com/litmus-cloud/prod/osx/litmus)
         - sha256: [https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256)

   - Run the deployment script with the following CLI command: 
     ```litmus deploy```

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

This project is not an official Google project. It is not supported by Google and Google specifically disclaims all warranties as to its quality, merchantability, or fitness for a particular purpose.

**Code Use and Cloud Costs:**

The code provided in this repository is provided "as is" without warranty of any kind, express or implied. It is your responsibility to understand the code, its dependencies, and its potential impact on your Google Cloud environment. 

Please be aware that deploying and running this application on Google Cloud will incur costs associated with the services it utilizes, such as Cloud Run, Firestore, and potentially others. You are solely responsible for monitoring and managing these costs. We recommend setting up appropriate budget alerts and monitoring tools within your Google Cloud Console to avoid unexpected expenses. 

**Security and Abuse:**

Also ensure you follow security best practices when deploying and configuring this application. Improper configuration or use could potentially lead to security vulnerabilities or abuse. We recommend reviewing the security documentation provided by Google Cloud and implementing appropriate security measures to protect your project.
