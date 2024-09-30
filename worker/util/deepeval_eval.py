# Copyright 2024 Google, LLC.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

"""This module contains functions for evaluating LLM responses using DeepEval."""

from google.cloud import logging
from deepeval import evaluate as deepeval
from deepeval.metrics import AnswerRelevancyMetric
from deepeval.models.base_model import DeepEvalBaseLLM
from deepeval.test_case import LLMTestCase
from langchain_google_vertexai import ChatVertexAI

from util.settings import settings


# Setup logging
logging_client = logging.Client()
WORKER_LOG_NAME = "litmus-worker-log"
worker_logger = logging_client.logger(WORKER_LOG_NAME)


# --- DeepEval Setup ---
# Base LLM for DeepEval
class GoogleVertexAIDeepEval(DeepEvalBaseLLM):
    """Class to implement Vertex AI for DeepEval"""

    def __init__(self, model):  # pylint: disable=W0231
        self.model = model

    def load_model(self):  # pylint: disable=W0221
        return self.model

    def generate(self, prompt: str) -> str:  # pylint: disable=W0221
        chat_model = self.load_model()
        return chat_model.invoke(prompt).content

    async def a_generate(self, prompt: str) -> str:  # pylint: disable=W0221
        chat_model = self.load_model()
        res = await chat_model.ainvoke(prompt)
        return res.content

    def get_model_name(self):  # pylint: disable=W0236 , W0221
        return "Vertex AI Model"


# Initialise safety filters for vertex model

generation_config = {"temperature": 0.0, "topk": 1}

# Initialize Gemini Pro for DeepEval
gemini_pro = ChatVertexAI(
    model_name="gemini-1.5-pro",
    generation_config=generation_config,
    project=settings.project_id,
    location=settings.location,
    response_validation=False,  # Important since deepeval cannot handle validation errors
)
deepeval_llm = GoogleVertexAIDeepEval(model=gemini_pro)
deepeval_metric = AnswerRelevancyMetric(
    threshold=0.5, model=deepeval_llm, async_mode=False
)


def evaluate_deepeval(question, answer, context):
    """Evaluates the LLM response using DeepEval.

    Args:
        actual_filtered_response (dict): The filtered response from the LLM.
        output_field (str): The key for the output field in the filtered response.
        golden_response (str): The expected (golden) response.

    Returns:
        dict or str: A dictionary containing the DeepEval evaluation results if successful,
                     or an error message if an exception occurs during evaluation.
    """
    try:
        # Create a test case
        deepeval_test_case = LLMTestCase(
            input=question,
            actual_output=answer,
            retrieval_context=[
                context
            ],  # You might need to provide context here based on your setup
        )

        # Evaluate with DeepEval and store results
        deepeval_metric.measure(deepeval_test_case)
        return {
            "metric": "AnswerRelevancy",
            "score": deepeval_metric.score,
            "reason": deepeval_metric.reason,
        }
    except Exception as e:
        worker_logger.log_text(
            f"Error in DeepEval evaluation: {str(e)}", severity="ERROR"
        )
        return f"Error during DeepEval evaluation: {str(e)}"
