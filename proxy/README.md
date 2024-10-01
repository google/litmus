## Litmus Proxy: Capture and Analyze Your LLM Interactions

The Litmus Proxy provides a powerful and transparent way to log and understand interactions with your Large Language Models (LLMs), including Google's Gemini family of models and other providers. By routing your LLM API traffic through the proxy, you gain valuable insights into usage patterns, performance, and potential areas for improvement.

**Benefits:**

- **Centralized Logging:** Capture all LLM API requests and responses in a unified log within BigQuery, simplifying monitoring and analysis.
- **Enhanced Debugging:** Easily troubleshoot issues and identify the root cause of unexpected behavior with detailed logs of LLM interactions.
- **Usage Analysis:** Gain insights into how your LLMs are being utilized, enabling you to optimize prompts, identify common queries, and track performance over time.
- **Contextualized Logging:** Associate logs with specific Litmus test runs by including the `X-Litmus-Request` header, providing deeper insights into LLM behavior within your testing workflows.
- **Customizable Integrations:** Forward logs to various destinations like Cloud Logging or your preferred monitoring tools for further analysis and visualization.

### Getting Started

#### Prerequisites

- Google Cloud Project with the following APIs enabled:
  - Cloud Run API
  - Secret Manager API
  - BigQuery API
- Google Cloud SDK installed and configured
- A BigQuery dataset for storing proxy logs (Litmus analytics setup creates this automatically)

#### Deployment

1. **Install the Litmus CLI:**

   - **Linux**:
     - install:`curl https://storage.googleapis.com/litmus-cloud/install/linux.sh | sudo sh`
     - binary: [https://storage.googleapis.com/litmus-cloud/prod/linux/litmus](https://storage.googleapis.com/litmus-cloud/prod/linux/litmus)
     - sha256: [https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/linux/litmus.sha256)
   - **OSX**:
     - install:`curl https://storage.googleapis.com/litmus-cloud/install/osx.sh | sudo sh`
     - binary: [https://storage.googleapis.com/litmus-cloud/prod/osx/litmus](https://storage.googleapis.com/litmus-cloud/prod/osx/litmus)
     - sha256: [https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256](https://storage.googleapis.com/litmus-cloud/prod/osx/litmus.sha256)

2. **Deploy the Proxy:**

   - Select a desired upstream URL from the available list by running:

   ```bash
   litmus proxy deploy
   ```

   - Alternatively, deploy with a specific upstream URL:

   ```bash
   litmus proxy deploy --upstreamURL <your_upstream_url>
   ```

   - Replace `<your_upstream_url>` with the desired upstream endpoint (e.g., `europe-west1-aiplatform.googleapis.com`).

#### Usage

1. **Retrieve Proxy Endpoint:**

   Obtain the URL of your deployed proxy service by running:

   ```bash
   litmus proxy list
   ```

   This command will display the names of your deployed proxy service.

2. **Configure your LLM Client (Python SDK):**

   Update your Vertex AI client initialization to utilize the proxy endpoint:

   **Before:**

   ```python
   import vertexai

   vertexai.init(project=project, location=location)
   ```

   **After:**

   ```python
   import vertexai

   proxy_endpoint = 'YOUR_PROXY_ENDPOINT' # Replace with the actual endpoint from step 1.

   vertexai.init(project=project, location=location, api_endpoint=proxy_endpoint, api_transport="rest")
   ```

3. **Optional: Adding Context**

   To associate proxy logs with specific Litmus test runs, you can add a context identifier to the proxy URL. The context is typically a unique ID associated with the test case in Litmus.

   Add `/litmus-context-{YOUR CONTEXT ID}` to the end of the proxy endpoint:

   **Context Example:**

   ```python
   import vertexai

   proxy_endpoint = 'YOUR_PROXY_ENDPOINT/litmus-context-{YOUR CONTEXT ID}'

   vertexai.init(project=project, location=location, api_endpoint=proxy_endpoint, api_transport="rest")

   ```

   **Note:** The `X-Litmus-Request` header will be automatically added to requests made through the Litmus worker service when you start a test run, so there is no need to add the header manually in those cases.

#### Additional Commands

- **Remove Proxy Service:**

  ```bash
  litmus proxy destroy <service_name>
  ```

- **Remove All Proxy Services:**

  ```bash
  litmus proxy destroy-all
  ```

### Log Analysis

The Litmus Proxy logs, stored in BigQuery, provide a comprehensive record of each LLM interaction, including:

- `id`: A UUID assigned to each log entry.
- `tracingID`: The value of the `X-Litmus-Request` header, enabling correlation with specific Litmus test runs.
- `litmusContext`: The context identifier extracted from the proxy URL, if present.
- `timestamp`: The timestamp of the request.
- `method`: The HTTP request method (e.g., POST).
- `requestURI`: The full request URI.
- `upstreamURL`: The upstream LLM endpoint the request was forwarded to.
- `requestHeaders`: The request headers, optionally excluding the `Authorization` header for security reasons.
- `requestBody`: The request body, parsed as JSON if possible.
- `requestSize`: The size of the request body in bytes.
- `responseStatus`: The HTTP response status code.
- `responseBody`: The response body, parsed as JSON if possible.
- `responseSize`: The size of the response body in bytes.
- `latency`: The request latency in milliseconds.

You can leverage these logs within BigQuery or the Litmus UI's Data Explorer to:

- **Monitor LLM Health and Performance:** Analyze latency trends, token usage, and response status codes to assess the overall health and performance of your LLMs.
- **Identify and Debug Issues:** Use detailed logs to pinpoint the root cause of errors or unexpected behavior in LLM responses, especially when correlated with specific Litmus test cases.
- **Analyze Usage Patterns and Optimize Prompts:** Gain insights into the most frequent requests, prompt structures, and parameter usage to optimize your LLM interactions for efficiency and cost-effectiveness.

### Customization

- **Authorization Header Logging:** By default, the proxy does not log the `Authorization` header for security reasons. You can enable this by setting the `LOG_AUTHORIZATION_HEADER` environment variable to `True` during proxy deployment.
- **Tracing Header:** The default tracing header is `X-Litmus-Request`. You can customize this by changing the `tracingHeader` variable in `main.go`. However, ensure consistency with your client and worker service configurations.

### Contribution

We encourage you to contribute your ideas and feedback to help us enhance the Litmus Proxy.
