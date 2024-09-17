# Getting Started

This guide will walk you through the initial steps of setting up Litmus and running your first tests.

<video controls="controls" src="/video/GettingStarted.mp4" />

## Prerequisites

**Before you start, ensure you have the following:**

- **Google Cloud Project:** You need an active Google Cloud Platform (GCP) project.
  - Create one: [https://cloud.google.com/resource-manager/docs/creating-managing-projects](https://cloud.google.com/resource-manager/docs/creating-managing-projects)
- **Billing Enabled:** Make sure billing is enabled for your project to utilize GCP services.
  - Enable billing: [https://cloud.google.com/billing/docs/how-to/modify-project](https://cloud.google.com/billing/docs/how-to/modify-project)
- **Google Cloud SDK (gcloud):** Install and configure the gcloud CLI to interact with your GCP project.
  - Download and install: [https://cloud.google.com/sdk/docs/install](https://cloud.google.com/sdk/docs/install)
  - Authenticate: `gcloud auth login`

## Quick Deployment with Litmus CLI

**The easiest way to set up Litmus is using the Litmus CLI.**

1. **Install the Litmus CLI:**

   - **Linux:**
     - Install:`curl https://storage.googleapis.com/litmus-cloud/install/linux.sh | sudo sh`
     - Binary: [https://storage.googleapis.com/litmus-cloud/prod/linux/litmus](https://storage.googleapis.com/litmus-cloud/prod/linux/litmus)
     - SHA256: [https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256)
   - **OSX:**
     - Install:`curl https://storage.googleapis.com/litmus-cloud/install/osx.sh | sudo sh`
     - Binary: [https://storage.googleapis.com/litmus-cloud/prod/osx/litmus](https://storage.googleapis.com/litmus-cloud/prod/osx/litmus)
     - SHA256: [https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256)

2. **Deploy Litmus:**

   ```bash
   litmus deploy
   ```

   - This will deploy the Litmus core services to your default GCP project.
   - The deployment script creates required service accounts, grants permissions, and deploys the worker and API services to Cloud Run.
   - You can customize the project and region using flags (see [CLI Usage](https://github.com/google/litmus/tree/main/cli)).

3. **Access Litmus:**
   - Run `litmus status` to retrieve the Litmus web interface URL and credentials.
   - Open the provided URL in your browser and log in using the displayed username and password.

## Next Steps

- **Create Test Templates:** Define and manage templates specifying the structure and parameters of your HTTP request tests.
- **Submit Test Runs:** Use the Litmus UI to submit test runs using pre-defined templates and provide test data.
- **(Optional) Deploy the Proxy Service:** For enhanced LLM monitoring, deploy the Litmus proxy service:

  ```bash
  litmus proxy deploy
  ```

  - This captures and logs interactions with your LLM (Vertex AI or other providers) to BigQuery for detailed analysis.

## Explore and Analyze

- **Web Interface:** Utilize the Litmus UI to monitor test run progress, view detailed results, filter responses for in-depth analysis, and gain insights into proxy logs and aggregated LLM usage metrics.

## Need Help?

- **Litmus CLI Usage:** Refer to the Litmus CLI documentation: [CLI Usage](/cli)
- **Proxy Service:** Learn about deploying and using the Litmus proxy service: [Proxy Usage](/proxy)
- **Contribute:** We welcome your contributions! See [Contribution Guide](/contribution) for details.

Let's get started with testing and evaluating your LLMs with Litmus!
