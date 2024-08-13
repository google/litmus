# Litmus CLI Deployment Script

This script automates the deployment of a Cloud Run service and a Cloud Run job with optional environment variables. It handles enabling necessary APIs, checking for existing Firestore databases, and deploying the services and jobs.

## Requirements

* Go 1.18 or later
* `gcloud` CLI installed and authenticated

## Usage

1. **Save the code:**  Save the provided code snippet as `deploy.go`.
2. **Build and run the script:**
   ```bash
   go build litmus.go
   ./deploy <project-id> [region] [env-var1=value1] [env-var2=value2] ...
   ```
   - Replace `<project-id>` with your Google Cloud project ID.
   - Optionally provide the region (defaults to "us-central1").
   - Add any environment variables you need in the format `env-var1=value1`.

## Example

```bash
 ./litmus deploy litmus-dev europe-west1 PASSWORD=test AI_DEFAULT_MODEL=gemini-1.5-flash 
```

This will deploy the Cloud Run services and jobs with the environment variables `DATABASE_URL` and `API_KEY` set to their respective values.

## Script Features

* **Automatic API Enabling:** Ensures required APIs (`artifactregistry`, `cloudbuild`, `run`, `firestore`) are enabled for your project.
* **Firestore Database Management:** Checks for the existence of a default Firestore database and creates one if it doesn't exist.
* **Environment Variable Injection:** Allows you to pass environment variable assignments as command-line arguments, which are automatically applied to both the service and the job.
* **Progress Indicator:** Displays a progress animation while the deployment is in progress.
* **URL Extraction:** Extracts and prints the URL of the deployed Cloud Run service.
* **Error Handling:** Includes logging and error handling to catch potential issues during deployment.

## Customization

* **Image Names:** Update the `--image` flags in the deployment commands to match your specific container images.
* **Service and Job Names:** Modify the service and job names used in the `gcloud` commands to reflect your desired names.
* **Environment Variables:** Adjust the environment variable assignments based on your application's requirements.
* **Flags:** Add or remove additional flags to the `gcloud` commands as needed for your specific deployment configuration.

## Notes

* Make sure your container images are pushed to a Google Container Registry (GCR) repository.
* This script assumes that your Cloud Run service and job are configured to use the default service account (which has access to the required APIs).
* For more complex deployment scenarios, consider using Google Cloud's Deployment Manager or Terraform.