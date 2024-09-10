# Adding Templates

This document outlines the steps to add a new test template using the Litmus User Interface (UI).

## Prerequisites

- Litmus deployed to Google Cloud, accessible through the web interface. See the main [Litmus README](https://github.com/google/litmus/blob/main/README.md) for deployment instructions.

## Steps

1. **Navigate to the Templates Page:** After logging into Litmus, click on "Templates" in the left-hand navigation menu.

2. **Open the Add Template Form:** Click the "Add Template" button. This will open a new form for creating a test template.

3. **Provide Template Details:**

   - **Template ID:** Enter a unique name to identify your template (e.g., "search-query-template").
   - **Test Cases:**
     - Click on "Add Test Case" to add individual test data items.
     - For each test case, provide:
       - **Query:** The input to your LLM (e.g., "What is the capital of France?").
       - **Response:** The expected output or golden answer from your LLM (e.g., "Paris").
       - **Filter** (optional): Any additional filters you want to apply (comma-separated).
       - **Source** (optional): A source identifier for the test case.
       - **Block** (optional): Toggle whether this test case should be blocked.
       - **Category** (optional): A category to organize your test cases.
   - **Request Payload:**
     - Use the JSON editor to define the structure of your test request.
     - Include placeholders (e.g., `{query}`) for dynamic values from your test cases.
   - **Pre-Request and Post-Request (optional):** Use the JSON editor to define optional requests to be executed before and after the main test request.
   - **LLM Evaluation Prompt (optional):** Provide a prompt to guide the LLM in assessing similarity between actual responses and golden responses.
   - **Input and Output Field Selection:**
     - Click on the "Input Field" button. The left-hand pane will display a JSON representation of your "Request Payload." Click the node corresponding to the field you wish to use as input to your test cases.
     - You can run your "Request Payload" to get an example response by clicking the "Test Request" button.
     - After running your request, click on the "Output Field" button. In the left-hand pane, click the node corresponding to the field you wish to use as the output for assessment.

4. **Test Your Request:** (Optional) Before saving, you can test the "Request Payload" by clicking the "Test Request" button. This ensures the request is valid and helps you visualize the response.

5. **Save the Template:** When finished, click the "Add Template" button. This will save your new template.

6. **Start a Run:** You can now use this template to submit test runs. Refer to the [Submitting Test Runs](/ui-start-test-run) documentation for details.

## Example JSON Upload

You can also populate test cases by uploading a JSON file. The JSON file must have an array with the following structure:

```json
[
  {
    "query": "What is the capital of France?",
    "response": "Paris",
    "filter": "location,city",
    "source": "wikipedia",
    "block": "false",
    "category": "geography"
  },
  {
    "query": "What is the highest mountain in the world?",
    "response": "Mount Everest",
    "filter": "mountain,height",
    "source": "nationalgeographic",
    "block": "false",
    "category": "geography"
  }
]
```

To upload your JSON:

- Go to the "Test Cases" tab.
- Click the "Upload JSON" button.
- Select your JSON file.

Litmus will validate the file and populate the test cases.
