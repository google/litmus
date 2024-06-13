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

from vertexai.preview.generative_models import (
    GenerativeModel,
    Tool,
    grounding,
    GenerationConfig,
)
import vertexai
from datetime import datetime
from zoneinfo import ZoneInfo
from util.docsnsnips import cleanup_json, strip_references
from util.settings import settings
import json

print("start assess")

vertexai.init(project=settings.project_id, location=settings.location)
model = GenerativeModel(settings.ai_default_model)
config = GenerationConfig(
    temperature=0.0, top_p=0.8, top_k=38, candidate_count=1, max_output_tokens=1000
)


def ask_llm_against_golden(statement, golden):
    """Calls the llm with custom prompt to compare a statement against the golden statement.
    Returns a dictionary with the results of this comparison. The result states whether the statement
    - is an answer at all
    - contradicts the golden statement
    - is equivalent (i.e. contains the same information as golden)
    - has additional information
    - is missing some information
    - is similar in structure to golden (0...1.0)
    Statements can be similar even if the information is contradictory (such as different figures for same topic).
    They can be dissimilar even if the information contained is the same.

    Args:
        statement (str): The statement to be assessed.
        golden (str): The golden response to compare against.

    Returns:
        dict: {
            'answered': bool,
            'contradictory': bool,
            'contradictory_explanation': str,
            'equivalent': bool,
            'equivalent_explanation': str,
            'addlinfo': bool,
            'addlinfo_explanation': str,
            'missinginfo': bool,
            'missinginfo_explanation': str,
            'similarity': float,
            'similarity_explanation': float
            }
    """
    # Prepare for prompting
    current_time = datetime.now(tz=ZoneInfo("Europe/Berlin"))

    llm_prompt = f"""
Today is {current_time.strftime('%A, %B %-d %Y')}. The current time is {current_time.strftime('%-H:%M')}.
You are a thorough quality inspector. Your task is to compare a statement about some topic to a golden response. The statement and the response can have different formats. Both statement and response are in German. You inspect the statement and the response to find out:
- has the question been answered at all?
- does the statement contradict the response?
- is the statement content-wise equivalent to the response, even it might have additional information?
- does the statement have additional information not contained in the response?
- is the statement missing information that is contained in the response?
- to what degree is the structure and wording of the statement similar to the response, even if content may be different?
Statements that are similar are estimated closer to 1 and statements that have different structure or wording are estimated closer to 0.
You MUST provide your output in JSON format. Do not provide any additional output.
This is what the JSON should look like:
{{
    "answered": 'true' if the question has been answered at all and 'false' if it has not,
    "contradictory": 'true' if the statement contradicts the response and 'false' if they agree,
    "contradictory_explanation": "explanation of how the statement contradicts the response if they don't agree",
    "equivalent": 'true' if the statement has equivalent information to the response and 'false' if the information differs,
    "equivalent_explanation": "explanation of how the two statements are different when they are not equivalent",
    "addlinfo": 'true' if the statement contains additional information compared to the response and 'false' if there is no additional information,
    "addlinfo_explanation": "explanation about the additional information if it present",
    "missinginfo": 'true' if the statement is missing information present in the response and 'false' if no information is missing,
    "missinginfo_explanation": "explanation about any missing information",
    "similarity": "provide a fractional numeric value between 0 and 1 that estimates the similarity of the statement to the response",
    "similarity_explanation": "explanation for the choice of value for the similarity attribute"
}}


Here is an example:
Statement: Der Fussballer B war 2011 und 2012 Fussballer des Jahres.
Golden response: 2010 und 2012 war B Fussballer des Jahres.

Comparison result:
{{
    "answered": true,
    "contradictory": true,
    "contradictory_explanation": "There is a contradiction because the years are different",
    "equivalent": false,
    "equivalent_explanation": "The years are different",
    "addlinfo": false,
    "addlinfo_explanation": "No additional information present",
    "missinginfo": false,
    "missinginfo_explanation": "No information is missing",
    "similarity": 0.8,
    "similarity_explanation": "The structure is similar but the facts are different"
}}


Here is another example:
Statement: Die Polizei im Kanton A wurde gestern wegen einer Unruhestörung zu einem Privathaus gerufen.
Best-known response: Aufgrund einer Unruhestörung rückte die Polizei gestern im Kanton A aus.

Comparison result:
{{
    "answered": true,
    "contradictory": false,
    "contradictory_explanation": "There is no contradiction",
    "equivalent": true,
    "equivalent_explanation": "Both statement and response mention the same incident",
    "addlinfo": true,
    "addlinfo_explanation": "The statement mentions the private house, the best-known response does not",
    "missinginfo": false,
    "missinginfo_explanation": "Nothing is missing",
    "similarity": 0.6,
    "similarity_explanation": "The information is similar but the wording is different"
}}


Here is another example:
Statement: C ist seit 2010 CEO von D.
Best-known response: C wurde 2010 zum CEO von D ernannt.

Comparison result:
{{
    "answered": true,
    "contradictory": false,
    "contradictory_explanation": "There is no contradiction",
    "equivalent": true,
    "equivalent_explanation": "Both statement and response mention the same facts",
    "addlinfo": false,
    "addlinfo_explanation": "No additional information present",
    "missinginfo": true,
    "missinginfo_explanation": "No information is missing",
    "similarity": 0.8,
    "similarity_explanation": "The structure is similar but the wording is different"
}}

Here is another example:
Statement: Diese Frage kann ich nicht beantworten.
Best-known response: Die Stadt X wurde 1833 gegründet.

Comparison result:
{{
    "answered": false,
    "contradictory": true,
    "contradictory_explanation": "There is a contradiction because there is a possible answer",
    "equivalent": false,
    "equivalent_explanation": "The question has not been answered",
    "addlinfo": false,
    "addlinfo_explanation": "No additional information present",
    "missinginfo": true,
    "missinginfo_explanation": "The facts from the golden response are missing",
    "similarity": 0,
    "similarity_explanation": "There is no answer provided"
}}

Here is your task:
Statement: {strip_references(statement)}
Best-known response: {strip_references(golden)}

Comparison result:
"""
    # Get LLM response
    # print('---------------------------')
    # print(LLM_PROMPT)
    try:
        responses = model.generate_content(
            llm_prompt, stream=False, generation_config=config
        )
        if responses.candidates[0].finish_reason == 1:
            result = responses.text
        else:
            result = "{'error': 'no response from LLM'}"
    except Exception as exc:
        print(exc)
        result = "{'error': 'exception calling LLM'}"
    comparison = cleanup_json(str(result))
    # print('---------------------------')
    # print(result)
    # print('---------------------------')
    # print(comparison)
    # print('---------------------------')
    try:
        comparison = json.loads(comparison)
    except:
        print(f"ERROR - llm response parsing failed for {comparison}")
        comparison = {"error": "llm response parsing failed"}
    return comparison


print("done assess")
