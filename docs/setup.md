# Setting Up Your Environment

This guide walks you through the detailed steps for setting up your Google Cloud environment and deploying Litmus manually.

![Litmus Architecture](/img/litmus-architecture.png)

## Prerequisites

- **Google Cloud Project:** An active GCP project with billing enabled.
- **Google Cloud SDK (gcloud):** Installed and configured with the correct project.
- **Required APIs:** Enable the following APIs in your GCP project:
  - **Cloud Run API**
  - **Firestore API**
  - **BigQuery API**
  - **Cloud Resource Manager API**
  - **Vertex AI API**
  - **Secret Manager API**

## 1. Google Cloud Setup

**1.1. Enable APIs:**

```bash
gcloud services enable run.googleapis.com firestore.googleapis.com bigquery.googleapis.com cloudresourcemanager.googleapis.com aiplatform.googleapis.com secretmanager.googleapis.com --project YOUR_PROJECT_ID
```

**1.2. Create Firestore Database:**

- Visit the Firestore console in your GCP project: [https://console.cloud.google.com/firestore](https://console.cloud.google.com/firestore)
- Create a database in a location of your choice.
- Select "Native mode".

**1.3. Create BigQuery Dataset:**

- Navigate to the BigQuery console: [https://console.cloud.google.com/bigquery](https://console.cloud.google.com/bigquery)
- Create a new dataset named "litmus_analytics" in your preferred location.

## 2. Service Accounts

**2.1. API Service Account:**

- In the IAM & admin console, create a service account named "litmus-api".
- Grant the following roles to the "litmus-api" service account at the project level:
  - **Vertex AI User:** Allows interaction with Vertex AI models.
  - **Cloud Datastore User:** Permits access to the Firestore database.
  - **Logs Writer:** Enables writing logs to Cloud Logging.
  - **Cloud Run Developer:** Allows deploying and managing Cloud Run services.
  - **BigQuery Data Viewer:** Grants read access to the BigQuery dataset.

**2.2. Worker Service Account:**

- Create another service account named "litmus-worker".
- Grant the same roles as in step 2.1 to the "litmus-worker" service account at the project level.
- Additionally, grant the "litmus-worker" service account the **Cloud Run Invoker** role on the "litmus-api" Cloud Run service. This allows the worker to invoke API endpoints to trigger test runs.

## 3. Password and Service URL Management

**3.1. Create Secret for Password:**

- Use Secret Manager to store the Litmus UI password securely:
  ```bash
  gcloud secrets create litmus-password --replication-policy=automatic --project YOUR_PROJECT_ID
  ```
- Add a secret version with a strong password (at least 16 characters):
  ```bash
  gcloud secrets versions add litmus-password --data-file=- <<EOF
  YOUR_STRONG_PASSWORD
  EOF
  ```

**3.2. Create Secret for Service URL:**

- Create another secret to store the Litmus API service URL (you'll populate this later):
  ```bash
  gcloud secrets create litmus-service-url --replication-policy=automatic --project YOUR_PROJECT_ID
  ```

## 4. Build and Deploy Docker Images

**4.1. Worker Service:**

- Build the worker service Docker image (from the `worker` directory):

  ```bash
  docker build -t gcr.io/YOUR_PROJECT_ID/litmus-worker:latest .
  ```

- Push the worker image to your Google Container Registry:

  ```bash
  docker push gcr.io/YOUR_PROJECT_ID/litmus-worker:latest
  ```

- Deploy the worker service as a Cloud Run _job_ (not a service) in your preferred region:
  ```bash
  gcloud run jobs deploy litmus-worker --image=gcr.io/YOUR_PROJECT_ID/litmus-worker:latest --region YOUR_REGION  --project YOUR_PROJECT_ID
  ```

**4.2. API Service:**

- Build the API service Docker image (from the `api` directory):

  ```bash
  docker build -t gcr.io/YOUR_PROJECT_ID/litmus-api:latest .
  ```

- Push the API image to your Google Container Registry:

  ```bash
  docker push gcr.io/YOUR_PROJECT_ID/litmus-api:latest
  ```

- Deploy the API service to Cloud Run:
  ```bash
  gcloud run deploy litmus-api --image gcr.io/YOUR_PROJECT_ID/litmus-api:latest --region YOUR_REGION  --allow-unauthenticated --project YOUR_PROJECT_ID
  ```

**4.3. Update Service URL Secret:**

- Once the API service is deployed, get its URL:

  ```bash
  gcloud run services describe litmus-api --region YOUR_REGION --format='value(status.url)' --project YOUR_PROJECT_ID
  ```

- Add this URL as a version to the "litmus-service-url" secret:
  ```bash
  gcloud secrets versions add litmus-service-url --data-file=- <<EOF
  YOUR_API_SERVICE_URL
  EOF
  ```

## 5. Deploy UI (Optional)

- The UI code is in the `api/ui` directory. You can host it using any web server (e.g., Nginx, Apache).
- Make sure the UI is configured to connect to the deployed API service URL (from the "litmus-service-url" secret).

## Using Litmus

1. **Access UI:** If you deployed the UI, navigate to its URL.
2. **Login:** Use "admin" as the username and the password you set in step 3.1.
3. **Create Templates:** Define your test request templates, including pre/post requests and LLM evaluation prompts.
4. **Submit Runs:** Initiate test runs using your templates.
5. **Analyze Results:** Explore and analyze the detailed results, including LLM-based assessments.

## Proxy Service Setup

- Refer to the [Proxy Service Documentation](/proxy) for detailed instructions on deploying and utilizing the Litmus proxy to monitor LLM interactions.

## Customization

- Modify environment variables during deployments to customize settings such as AI models, regions, and authentication (see `api/util/settings.py`).

This comprehensive guide will help you set up and utilize Litmus for evaluating and testing your LLMs.
