# Evaluating LLMs with Custom LLM Evaluation

This document outlines how to enable, use, and configure Litmus's custom LLM evaluation feature, which leverages the power of LLMs to assess the quality and correctness of other LLMs' responses.

## What is Custom LLM Evaluation?

Custom LLM Evaluation allows you to harness an LLM to compare an LLM's output (statement) against a "golden" response, providing a more nuanced evaluation than simple string matching. This is particularly useful when dealing with complex responses or when the definition of correctness is subjective.

## Enabling Custom LLM Evaluation

1. **Edit your test template:**
   - Navigate to the "Templates" page in the Litmus UI and click the "Edit" button next to the desired template.
2. **Enable Custom LLM Evaluation in the "LLM Evaluation Prompt" tab:**
   - Check the checkbox for **Custom LLM Evaluation**.
3. **(Optional) Customize the LLM Evaluation Prompt:**
   - You can provide a custom prompt to guide the evaluating LLM. By default, Litmus includes a comprehensive prompt that covers aspects like:
     - Has the question been answered?
     - Are the statement and golden response contradictory?
     - Are they equivalent in terms of content?
     - Does the statement contain additional or missing information?
     - How similar are the structure and wording of the two statements?
   - You can modify this prompt or write your own based on your specific evaluation requirements.
4. **Select the Input and Output Fields:**
   - In the same "LLM Evaluation Prompt" tab:
     - **Input Field:** Click this button to select the field in the "Request Payload" that serves as the input to the LLM being tested.
     - **Output Field:** Click this button to select the field in the "Response Payload" that will be assessed against the golden response.
     - You'll need to run a test request to obtain an example response and see the available output fields.
5. **Save your template:**
   - Click the "Update Template" button to save the changes.

## Using Custom LLM Evaluation

Once enabled, Litmus will execute the LLM assessment for every test case in a run that utilizes the modified template. The results will be included within the "assessment" field of the corresponding test case:

```json
{
  "status": "Passed",
  "response": {
    // ... actual response data ...
  },
  "assessment": {
    "llm_assessment": {
      "answered": true,
      "contradictory": false,
      "contradictory_explanation": null,
      "equivalent": true,
      "equivalent_explanation": "The statement and golden response convey the same information.",
      "addlinfo": true,
      "addlinfo_explanation": "The statement provides an additional detail about...",
      "missinginfo": false,
      "missinginfo_explanation": null,
      "similarity": 0.9,
      "similarity_explanation": "The statements are highly similar in structure and wording."
    }
  }
}
```

The "status" field of the test case will be updated based on the `similarity` score returned by the LLM. If the similarity is above 0.5, the status will be "Passed"; otherwise, it will be "Failed".

## Important Notes

- **Prompt Engineering:** Carefully craft your LLM evaluation prompt to ensure it provides clear and specific instructions for the assessment.
- **Bias Awareness:** Be mindful of potential biases in the evaluating LLM. Evaluate its performance and consider mitigating biases through techniques like prompt engineering.
- **Resource Management:** Using LLMs for evaluation consumes tokens. Monitor token usage and consider strategies to optimize prompts and manage costs.

By effectively utilizing custom LLM evaluation, you can create more sophisticated and context-aware assessments for your LLM applications, leading to more reliable and insightful testing results.
