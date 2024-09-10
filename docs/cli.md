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
  status      Show the status of the Litmus deployment
  version     Display the version of the Litmus CLI
  execute     Execute a payload against the Litmus application
  ls          List all runs
  run         Open a specific Litmus run
  analytics   Manage Litmus analytics (deploy or destroy)
  proxy       Manage Litmus proxy (deploy, list, destroy, destroy-all)

Options:
  --project <project-id>: Specify the project ID (overrides default)
  --region <region>: Specify the region (defaults to 'us-central1')
  --quiet                Suppress verbose output

```

### Examples

- **Deploy Litmus:**

  ```bash
  litmus deploy
  ```

- **Deploy to a specific project and region:**

  ```bash
  litmus deploy --project my-project-id --region us-east1
  ```

- **Destroy the Litmus deployment:**

  ```bash
  litmus destroy
  ```

- **Get deployment status:**

  ```bash
  litmus status
  ```

- **Display CLI version:**

  ```bash
  litmus version
  ```

- **Execute a payload:**

  ```bash
  litmus execute "Hello, world!"
  ```

- **List all runs:**

  ```bash
  litmus ls
  ```

- **Open a specific run:**

  ```bash
  litmus run <runID>
  ```

- **Deploy Litmus Analytics:**

  ```bash
  litmus analytics deploy
  ```

- **Destroy the Litmus Analytics deployment:**

  ```bash
  litmus analytics destroy
  ```

- **Deploy Litmus Proxy:**

  ```bash
  litmus proxy deploy
  ```

  This will prompt the user to select from a list of available regions and platforms. Alternatively, the upstream URL can be specified:

  ```bash
  litmus proxy deploy --upstreamURL <your_upstream_url>
  ```

  - Replace `<your_upstream_url>` with the desired upstream endpoint (e.g., `europe-west1-aiplatform.googleapis.com`).

- **List all deployed Litmus Proxies:**

  ```bash
  litmus proxy list
  ```

- **Destroy a Litmus Proxy deployment:**

  ```bash
  litmus proxy destroy <service_name>
  ```

  - Replace `<service_name>` with the name of the deployed proxy (e.g., `us-central1-aiplatform-litmus-abcd`).

- **Destroy all Litmus Proxy deployments:**
  ```bash
  litmus proxy destroy-all
  ```

## Configuration

- The CLI uses your default gcloud project configuration.
- Use the `--project` flag to specify a different project.
