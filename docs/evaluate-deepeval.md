# Evaluating LLMs with DeepEval

This document provides a comprehensive guide to enabling, using, configuring, and extending DeepEval within the Litmus framework for evaluating LLM responses.

## What is DeepEval?

DeepEval is a Python library specifically designed for evaluating the quality of responses generated by LLMs. It leverages other LLMs to perform the evaluations, offering a more nuanced and human-like assessment compared to traditional metric-based approaches. DeepEval supports a variety of metrics, including:

- **Answer Relevancy:** Measures how well the answer relates to the question.
- **Faithfulness:** Assesses the factual accuracy of the answer in relation to the provided context.
- **Contextual Precision:** Evaluates the conciseness and relevance of the answer by measuring how much of the retrieved context is actually used.
- **Contextual Recall:** Measures how well the answer covers the relevant information from the given context.
- **Contextual Relevancy:** Assesses the overall relevance of the answer, considering both the question and the context.
- **Hallucination:** Checks for information in the answer that is not supported by the provided context.
- **Bias:** Evaluates potential biases in the answer.
- **Toxicity:** Checks for harmful or offensive language in the answer.

## Enabling DeepEval in Litmus

DeepEval is disabled by default in Litmus. To enable it, you need to modify your test templates:

1. **Edit your test template:**
   - In the Litmus UI, navigate to the "Templates" page and click the "Edit" button next to the template you want to modify.
2. **Enable DeepEval in the "LLM Evaluation Prompt" tab:**
   - Check the checkbox for **DeepEval**. This will reveal a list of available DeepEval metrics.
3. **Select DeepEval metrics:**
   - Choose the specific DeepEval metrics you want to use for evaluation. You can select multiple metrics.
4. **Save your template:**
   - Click the "Update Template" button to save your changes.

## Using DeepEval

Once DeepEval is enabled and you've selected the desired metrics, Litmus will automatically apply these evaluations to LLM responses during test runs. The results are embedded within the `assessment` field of the test case, with each selected metric having its own subfield:

```json
{
  "status": "Passed",
  "response": {
    "output": "This is the answer"
  },
  "assessment": {
    "answer_relevancy_deepeval_evaluation": {
      "metric": "AnswerRelevancyMetric",
      "score": 0.9,
      "reason": "The answer is relevant to the question."
    },
    "faithfulness_deepeval_evaluation": {
      "metric": "FaithfulnessMetric",
      "score": 0.8,
      "reason": "The answer is mostly faithful to the context."
    }
    // ... other selected DeepEval metrics
  }
}
```

```

## Configuring and Extending DeepEval

The configuration and extension of DeepEval, like Ragas, **require code changes within the worker service.**

- **Adding New Metrics:** To incorporate new DeepEval metrics, you'd modify the `deepeval_metric_factory` function in `deepeval_eval.py` within the worker service code. This involves adding new mappings from metric names to their corresponding DeepEval metric classes.
- **Adjusting Thresholds:** To change the default pass/fail thresholds for DeepEval metrics, you'd modify the threshold values when instantiating the metric objects within `deepeval_metric_factory`.

**Note:** Remember that modifications to the worker service code require rebuilding and redeploying the worker Docker image for the changes to take effect.

## Important Notes

- Ensure your test cases provide sufficient context for DeepEval to assess faithfulness and other context-dependent metrics accurately.
- Be aware that DeepEval uses another LLM for evaluation. This introduces another layer of complexity and potential biases.
- Stay updated with DeepEval's development and consider incorporating new metrics or features as they become available.
```