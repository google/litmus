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

"""This module defines the function ask_llm_against_golden() which calls a large language model (LLM) 
to compare two statements. The function is intended for use in applications that need to assess 
the similarity and consistency of different pieces of text, such as comparing user responses 
to expected answers or evaluating the quality of generated text."""

import vertexai
from vertexai.generative_models import GenerativeModel, GenerationConfig
from util.settings import settings

# Initialize Vertex AI with project and location settings
vertexai.init(project=settings.project_id, location=settings.region)

# Load the specified LLM
model = GenerativeModel(settings.ai_validation_model)

# Configure generation parameters for the LLM
config = GenerationConfig(
    temperature=0.0,  # Control the randomness of the generated text (0.0 = deterministic)
    top_p=0.8,  # Control the diversity of the generated text
    top_k=38,  # Limit the vocabulary size for generation
    candidate_count=1,  # Generate only one candidate response
    max_output_tokens=8000,  # Set the maximum number of tokens for the generated response
)


def ask_llm_for_summary(**kwargs):
    """Generates an executive summary using an LLM."""

    if "summaries" in kwargs:
        prompt = f"Generate an overall executive summary for the following test summaries:\n\n{kwargs.get('summaries')} \n Please remember to answer in full sentences and give me cohesive paragraphs."
    elif "outliers" in kwargs:
        prompt = f"Give me the outliers for the following test summaries:\n\n{kwargs.get('outliers')} \n Please remember to answer in full sentences and give me cohesive paragraphs."
    else:
        prompt = f"Generate an executive summary for the following test case:\n\nQuestion:{kwargs.get('question')}\n\nAnswer:{kwargs.get('answer')}\n\nGolden Response:{kwargs.get('golden')} \n Please remember to answer in full sentences and give me cohesive paragraphs. Please also state in the beginning what the input text, the output text and the expected result was."

    response = model.generate_content(prompt, generation_config=config)
    return response.text
