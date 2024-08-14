# Litmus CLI

A command-line interface for deploying and managing Litmus, a tool for quickly building and testing LLMs.

## Prerequisites

- **Google Cloud SDK (gcloud)**: Ensure you have the Google Cloud SDK installed and authenticated.
  - Install: [https://cloud.google.com/sdk/docs/install](https://cloud.google.com/sdk/docs/install)
  - Authenticate: `gcloud auth login`
- **Go 1.18 or higher**: Required for building and running the CLI.

## Installation

1. Clone this repository to your local machine:

   ```bash
   git clone https://github.com/your-username/litmus.git
   ```

2. Navigate to the project directory:

   ```bash
   cd litmus
   ```

3. Build the CLI:

   ```bash
   go build
   ```

## Usage

```
Usage: go run main.go <command> [options] [flags] 

Commands:
  open: Open the Web application in your browser
  deploy: Deploy the application
  destroy: Remove the application
  status: Show the status of the Litmus deployment
  version: Display the version of the Litmus CLI
  execute <payload>: Execute a payload to the deployed endpoint
  ls: List all runs
  run <runID>: Open the URL for a certain runID 

Options:
  --project <project-id>: Specify the project ID (overrides default)
  --region <region>: Specify the region (defaults to 'us-central1')
```

### Examples

- **Deploy Litmus:**
  ```bash
  go run main.go deploy
  ```

- **Deploy to a specific project and region:**
  ```bash
  go run main.go deploy --project my-project-id --region us-east1
  ```

- **Destroy the Litmus deployment:**
  ```bash
  go run main.go destroy 
  ```

- **Get deployment status:**
  ```bash
  go run main.go status 
  ```

- **Display CLI version:**
  ```bash
  go run main.go version
  ```

- **Execute a payload:**
  ```bash
  go run main.go execute "Hello, world!"
  ```

- **List all runs:**
  ```bash
  go run main.go ls 
  ```

- **Open a specific run:**
  ```bash
  go run main.go run <runID>
  ```

## Configuration

- The CLI uses your default gcloud project configuration.
- Use the `--project` flag to specify a different project.