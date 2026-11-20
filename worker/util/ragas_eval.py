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

"""This module contains functions for evaluating LLM responses using RAGAS."""

from google.cloud import logging
from datasets import Dataset
from ragas.llms.base import LangchainLLMWrapper
from ragas import evaluate
from ragas.metrics import (
    answer_relevancy,
    answer_similarity,
    context_precision,
    context_recall,
)
from ragas.metrics.critique import harmfulness
from langchain_google_vertexai import VertexAI, VertexAIEmbeddings

from util.settings import settings


# Setup logging
logging_client = logging.Client()
WORKER_LOG_NAME = "litmus-worker-log"
worker_logger = logging_client.logger(WORKER_LOG_NAME)


# Load Gemini Pro model and embeddings for RAGAS
gemini_pro = VertexAI(
    model_name="gemini-1.5-pro", project=settings.project_id, location=settings.location
)
embeddings = VertexAIEmbeddings(
    model_name="textembedding-gecko@003",
    project=settings.project_id,
    location=settings.location,
)

# Compile list of RAGAS Metrics
ragas_metrics = [
    answer_relevancy,
    context_recall,
    context_precision,
    harmfulness,
    answer_similarity,
]


# IMPORTANT: Gemini with RAGAS
# RAGAS is designed to work with OpenAl Models by default. We must set a few attributes to make it work with Gemini
class RAGASVertexAIEmbeddings(VertexAIEmbeddings):
    """Wrapper for RAGAS"""

    async def embed_text(self, text: str) -> list[float]:
        """Embeds a text for semantics similarity"""
        return self.embed([text], 1, "SEMANTIC_SIMILARITY")[0]


# Wrapper to make RAGAS work with Gemini and Vertex AI Embeddings Models
ragas_embeddings = RAGASVertexAIEmbeddings(
    model_name="textembedding-gecko@003",
    project=settings.project_id,
    location=settings.location,
)
ragas_llm = LangchainLLMWrapper(gemini_pro)
for m in ragas_metrics:
    # change LLM for metric
    m.__setattr__("llm", ragas_llm)
    # check if this metric needs embeddings
    if hasattr(m, "embeddings"):
        # if so change with Vertex AI Embeddings
        m.__setattr__("embeddings", ragas_embeddings)


def evaluate_ragas(question, answer, golden_response, context):
    """Evaluates the LLM response using RAGAS metrics.

    Args:
        actual_filtered_response (dict): The filtered response from the LLM.
        output_field (str): The key for the output field in the filtered response.
        golden_response (str): The expected (golden) response.

    Returns:
        dict or str: A dictionary containing the RAGAS evaluation results if successful,
                     or an error message if an exception occurs during evaluation.
    """

    try:
        # Convert to a dataset
        ragas_dataset = Dataset.from_dict(
            {
                "question": [
                    question,
                ],
                "answer": [
                    answer,
                ],
                "contexts": [
                    [
                        context,
                    ],
                ],
                "ground_truth": [
                    golden_response,
                ],
            }
        )

        # Evaluate with RAGAS and store results
        ragas_result = evaluate(
            ragas_dataset,
            metrics=ragas_metrics,
            raise_exceptions=False,
        )

        return ragas_result.to_pandas().to_dict()

    except Exception as e:
        worker_logger.log_text(f"Error in RAGAS evaluation: {str(e)}", severity="ERROR")
        return f"Error during RAGAS evaluation: {str(e)}"
