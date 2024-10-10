# Litmus CLI

A command-line interface for deploying and managing Litmus, a tool for quickly building and testing LLMs.

## Prerequisites

- **Google Cloud SDK (gcloud)**: Ensure you have the Google Cloud SDK installed and authenticated.
  - Install: [https://cloud.google.com/sdk/docs/install](https://cloud.google.com/sdk/docs/install)
  - Authenticate: `gcloud auth login`
- **Go 1.18 or higher**: Required for building and running the CLI.

## Installation

### Fast installation

Make sure you have the Google Cloud SDK installed and configured with the correct project.

Install binary:

- **Linux**:
  `curl https://storage.googleapis.com/litmus-cloud/install/linux.sh | sudo sh`
- sha256: [https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256)
- **OSX**:
  `curl https://storage.googleapis.com/litmus-cloud/install/osx.sh | sudo sh`
- sha256: [https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256)

### Manual Build

1. Clone this repository to your local machine:

   ```bash
   git clone https://github.com/google/litmus.git
   ```

2. Navigate to the project directory:

   ```bash
   cd litmus/cli
   ```

3. Install Go dependencies:

   ```bash
   go mod download
   ```

4. Build the CLI:

   ```bash
   go build
   ```

## Usage

```
Usage: litmus <command> [options] [flags]

Commands:

  open        Open the Litmus dashboard
  deploy      Deploy the application
  destroy     Destroy Litmus resources
  update      Update the application
  status      Show the status of the Litmus deployment
  version     Display the version of the Litmus CLI
  execute     Execute a payload against the Litmus application
  ls          List all runs
  run         Open a specific Litmus run
  start       Starts a new Litmus run
  analytics   Manage Litmus analytics (deploy or destroy)
  proxy       Manage Litmus proxy (deploy, list, destroy, destroy-all)

Options:
  --project <project-id>: Specify the project ID (overrides default)
  --region <region>: Specify the region (defaults to 'us-central1')
  --quiet                Suppress verbose output
  --preserve-data        Preserve data in Cloud Storage, Firestore, and BigQuery

```

### Examples

- **Deploy Litmus:**

  ```bash
  litmus deploy
  ```

  This command deploys the Litmus core services (API and Worker) to your default GCP project in the `us-central1` region. During deployment it will create required service accounts, grant permissions and deploy the services to Cloud Run. You can use the `--quiet` flag to suppress verbose output.

- **Deploy to a specific project and region:**

  ```bash
  litmus deploy --project my-project-id --region us-east1
  ```

  This command deploys the Litmus core services (API and Worker) to the specified GCP project (`my-project-id`) in the specified region (`us-east1`).

- **Deploy to a specific environment:**

  ```bash
  litmus deploy dev
  ```

  This command deploys the Litmus core services to the `dev` environment. This will pull and deploy the latest `dev` images.

- **Destroy the Litmus deployment:**

  ```bash
  litmus destroy
  ```

  This command deletes all Litmus resources in your default project and `us-central1` region. It removes the API and worker service deployments, deletes secrets from secret manager, service accounts, and the Cloud Storage bucket. You can use the `--quiet` flag to suppress verbose output.

- **Destroy the Litmus deployment and preserve data:**

  ```bash
  litmus destroy --preserve-data
  ```

  This command deletes all Litmus resources in your default project and `us-central1` region but keeps the data in Cloud Storage, Firestore and BigQuery.

- **Update the Litmus deployment:**

  ```bash
  litmus update
  ```

  This command updates your Litmus deployment to the latest version available. It updates both the API and the Worker deployments. You can use the `--quiet` flag to suppress verbose output.

- **Update the Litmus deployment to a specific environment:**

  ```bash
  litmus update dev
  ```

  This command updates your Litmus deployment to the latest `dev` version available. It updates both the API and the Worker deployments.

- **Get deployment status:**

  ```bash
  litmus status
  ```

  This command retrieves and displays the status of your Litmus deployment. This includes the service URL, username and password.

- **Display CLI version:**

  ```bash
  litmus version
  ```

  This command displays the version of the installed Litmus CLI.

- **Execute a payload:**

  ```bash
  litmus execute "Hello, world!"
  ```

  This is a placeholder command, there is no implementation yet.

- **List all runs:**

  ```bash
  litmus ls
  ```

  This command retrieves and displays a list of all the test runs that have been submitted, including their status and other details.

- **Open a specific run:**

  ```bash
  litmus run <runID>
  ```

  This command opens the details page for a specific Litmus run in your default browser. You can get the run ID by running the `litmus ls` command.

- **Start a new Litmus Test Run:**

  ```bash
  litmus start $TEMPLATE_ID $RUN_ID
  ```

  This command submits a new test run using the provided template ID and run ID. Make sure that the template exists before running the command. The `$RUN_ID` can be generated automatically by running `uuidgen`.

- **Deploy Litmus Analytics:**

  ```bash
  litmus analytics deploy
  ```

  This command sets up the analytics components for Litmus, including a BigQuery dataset for storing logs and log sinks to route logs from the proxy and API to BigQuery.

- **Destroy the Litmus Analytics deployment:**

  ```bash
  litmus analytics destroy
  ```

  This command removes the analytics components for Litmus, including the BigQuery dataset and the log sinks.

- **Deploy Litmus Proxy:**

  ```bash
  litmus proxy deploy
  ```

  This command deploys the Litmus proxy service. It will prompt the user to select from a list of available regions and platforms. The Proxy is used for logging and analyzing your LLM interactions.

- **Deploy Litmus Proxy for specific upstream URL:**

  ```bash
  litmus proxy deploy --upstreamURL <your_upstream_url>
  ```

  This command deploys the Litmus proxy service for a specific upstream URL. Replace `<your_upstream_url>` with the desired upstream endpoint (e.g., `europe-west1-aiplatform.googleapis.com`).

- **List all deployed Litmus Proxies:**

  ```bash
  litmus proxy list
  ```

  This command lists all Litmus proxy services that are currently deployed in your GCP project. It displays the name and URL of each proxy.

- **Destroy a Litmus Proxy deployment:**

  ```bash
  litmus proxy destroy <service_name>
  ```

  This command destroys the specified Litmus proxy deployment. Replace `<service_name>` with the name of the deployed proxy (e.g., `us-central1-aiplatform-litmus-abcd`).

- **Destroy all Litmus Proxy deployments:**
  ```bash
  litmus proxy destroy-all
  ```
  This command destroys all Litmus proxy deployments in your current project and region.

## Configuration

- The CLI uses your default gcloud project configuration.
- You can use the `--project` flag to specify a different project for all commands.
- You can use the `--region` flag to specify a different region for the `deploy` and `destroy` commands.
- Most commands accept environment variables as key-value pairs separated by `=`, for example:
  ```bash
  litmus deploy KEY1=VALUE1 KEY2=VALUE2
  ```
