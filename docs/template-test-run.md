# Test Run Template

The "Test Run" template in Litmus is designed for executing a set of independent test cases against your AI model, primarily focusing on single-turn interactions where you provide an input and expect a specific output. Each test case is treated as an isolated unit, and the results are evaluated individually.

## Structure

A "Test Run" template in Litmus consists of the following key components:

- **Template ID:** A unique identifier for your template.
- **Template Data:** An array of test cases, each containing:
  - **Query:** The input to your AI model.
  - **Response:** The expected output or "golden answer" from your AI model.
  - **Filter** (optional): Comma-separated keywords or categories to organize and filter test cases.
  - **Source** (optional): A source identifier for the test case.
  - **Block** (optional): A boolean (true/false) indicating whether this test case should be excluded from the run.
  - **Category** (optional): A category or label for the test case.
- **Request Payload:** A JSON object defining the structure of the request sent to the AI model. It includes placeholders for dynamic values from your test cases, typically using curly braces notation (e.g., `{query}`).
- **Pre-Request and Post-Request (optional):** JSON objects defining optional requests to be executed before and after the main test request, respectively. These are useful for setting up or cleaning up your testing environment.
- **LLM Evaluation Prompt (optional):** A prompt guiding the LLM in assessing the similarity between the actual responses and the golden responses. This allows for a more nuanced evaluation beyond simple equality checks.
- **Input and Output Field Selection:** Specifies the exact fields within the request and response payloads to use as input and output for the test cases and LLM assessment.

## Purpose

The primary purpose of a "Test Run" template is to:

- **Automate testing of various inputs:** Efficiently run numerous test cases against your AI model without manual intervention.
- **Verify expected behavior:** Ensure your model responds as intended for a range of input queries.
- **Identify inconsistencies and regressions:** Detect deviations from the expected outputs, highlighting potential issues or regressions in your model.
- **Facilitate model improvement:** Use the results to refine your model's training data or adjust its parameters.

## Scenarios

### Good Use Cases

- **Unit testing individual model functionalities:** Verify the accuracy and consistency of specific model capabilities, such as question answering, text summarization, or translation.
- **Regression testing after model updates:** Ensure that changes to your model haven't introduced unintended consequences or broken existing functionalities.
- **Benchmarking model performance against different datasets:** Compare your model's performance on various test sets to identify strengths and weaknesses.

### Bad Use Cases

- **Multi-turn conversations or dialogues:** "Test Run" templates are not suitable for testing conversational flows as each test case is independent and lacks context from previous turns. Use "Test Missions" for this purpose.
- **Testing model behavior that requires state or context:** If your model's output depends on previous interactions or user-specific information, "Test Run" templates might not be appropriate.

## Starting a Test Run

### API

1. Construct a JSON payload with the following information:
   ```json
   {
     "run_id": "your-unique-run-id",
     "template_id": "your-template-id",
     "pre_request": { ... }, // Optional
     "post_request": { ... }, // Optional
     "test_request": { ... }
   }
   ```
2. Send a POST request to the `/submit_run` endpoint of the Litmus API.

### CLI (Advanced)

1. Gather the following information:
   - `RUN_ID`: A unique identifier for your test run.
   - `TEMPLATE_ID`: The ID of your "Test Run" template.
   - `GCP_PROJECT`: Your Google Cloud project ID.
   - `GCP_REGION`: The region where your Litmus worker is deployed.
   - Optional: Prepare JSON strings for `PRE_REQUEST` and `POST_REQUEST` if needed.
2. Use the `gcloud` CLI to invoke the Litmus worker job with the gathered information as environment variables:
   ```bash
   gcloud run jobs execute litmus-worker \
       --project $GCP_PROJECT \
       --region $GCP_REGION \
       --set-env-vars RUN_ID=$RUN_ID,TEMPLATE_ID=$TEMPLATE_ID,PRE_REQUEST='$PRE_REQUEST',POST_REQUEST='$POST_REQUEST'
   ```

### CLI (Simple)

1. Get your `RUN_ID` and `TEMPLATE_ID`.
2. Run the following `litmus` CLI command:
   ```bash
   litmus start $TEMPLATE_ID $RUN_ID
   ```

### UI

1. Navigate to the "Start New Run" page.
2. Select your "Test Run" template from the dropdown.
3. Enter your `RUN_ID`.
4. Review and modify the request payload if needed.
5. Submit the run.

## Configuration

You can configure various aspects of a "Test Run" template in the UI, including:

- Modifying the `Request Payload` to match your API requirements.
- Defining or updating the `Pre-Request` and `Post-Request`.
- Crafting the `LLM Evaluation Prompt` to guide the assessment process.
- Selecting the appropriate `Input Field` and `Output Field`.

## Restarting and Deleting

- **Restart:** You can restart a "Test Run" by sending a POST request to the `/invoke_run` endpoint with the `RUN_ID` and `TEMPLATE_ID` or by using the UI's restart button.
- **Delete:** Delete a "Test Run" by sending a DELETE request to the `/delete_run/<run_id>` endpoint or by using the UI's delete button.

## Additional Information

- The results of each test case, including the LLM assessment, are stored in Firestore.
- You can filter and analyze test run results using the UI or by querying the Firestore database directly.
- Consider organizing your test cases using the optional `Filter` and `Category` fields for better management and analysis.
