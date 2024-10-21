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
from datetime import datetime
from zoneinfo import ZoneInfo
import json

from util.docsnsnips import cleanup_json, strip_references
from util.settings import settings

print("Initializing assessment module...")

# Initialize Vertex AI with project and location settings
vertexai.init(project=settings.project_id, location=settings.location)

# Load the specified LLM
model = GenerativeModel(settings.ai_default_model)

# Configure generation parameters for the LLM
config = GenerationConfig(
    temperature=0.0,  # Control the randomness of the generated text (0.0 = deterministic)
    top_p=0.8,  # Control the diversity of the generated text
    top_k=38,  # Limit the vocabulary size for generation
    candidate_count=1,  # Generate only one candidate response
    max_output_tokens=1000,  # Set the maximum number of tokens for the generated response
)


def ask_llm_against_golden(statement, golden, prompt):
    """
    Compares a statement against a golden statement using a large language model (LLM).

    This function sends a prompt to the LLM, asking it to compare the given statement
    against a "best-known" (golden) statement. The LLM's response is then parsed to
    extract information about the comparison, such as whether the statements are
    contradictory, equivalent, or share similarities.

    Args:
        statement (str): The statement to be assessed.
        golden (str): The golden response to compare against.
        prompt (str): The prompt to guide the LLM's comparison.

    Returns:
        dict: A dictionary containing the results of the comparison, including:
            - 'answered': Whether the statement is a valid response.
            - 'contradictory': Whether the statement contradicts the golden statement.
            - 'contradictory_explanation': Explanation of the contradiction.
            - 'equivalent': Whether the statement is equivalent to the golden statement.
            - 'equivalent_explanation': Explanation of the equivalence.
            - 'addlinfo': Whether the statement contains additional information.
            - 'addlinfo_explanation': Explanation of the additional information.
            - 'missinginfo': Whether the statement is missing information.
            - 'missinginfo_explanation': Explanation of the missing information.
            - 'similarity': Similarity score between the statements (0-1).
            - 'similarity_explanation': Explanation of the similarity score.
            - 'error': Error message if something goes wrong.
    """

    # Get current time in Berlin timezone
    current_time = datetime.now(tz=ZoneInfo("Europe/Berlin"))

    # Construct the prompt for the LLM
    llm_prompt = f"""
Today is {current_time.strftime('%A, %B %-d %Y')}. The current time is {current_time.strftime('%-H:%M')}.

{prompt}

Here is your task:
Statement: {strip_references(statement)}
Best-known response: {golden}

Comparison result:
"""

    try:
        # Send the prompt to the LLM and get the response
        responses = model.generate_content(
            llm_prompt, stream=False, generation_config=config
        )

        # Check if the LLM finished generating the response successfully
        if responses.candidates[0].finish_reason == 1:
            result = responses.text
        else:
            result = "{'error': 'no response from LLM'}"
    except Exception as exc:
        print(f"Error calling LLM: {exc}")
        result = "{'error': 'exception calling LLM'}"

    # Clean up the LLM response and parse it as JSON
    comparison = cleanup_json(str(result))

    try:
        comparison = json.loads(comparison)
    except:
        print(f"ERROR - llm response parsing failed for {comparison}")
        comparison = {"error": "llm response parsing failed"}

    return comparison


def ask_llm_for_action(mission_description, conversation_history):
    """Asks the LLM for the next action to take in the test mission.

    Args:
        mission_description (str): The description of the overall mission.
        conversation_history (list): A list of previous turns in the conversation.

    Returns:
        dict: A dictionary containing the LLM's suggested action, including:
            - 'request' (dict): Data for the API request to be made.
              Should include 'url', 'method', 'body' (if applicable), and 'headers'.
    """

    # Construct the conversation history string for the prompt
    conversation_string = "\n".join(
        [
            f"{turn.get('role', 'unknown')}: {turn['content']}"
            for turn in conversation_history
        ]
    )

    # Build the prompt for the LLM
    prompt = f"""You are helping to complete a test mission. 
    The mission description is: {mission_description}

    Here is the conversation history:
    {conversation_string}

    What is the next request that should be sent?
    Respond with a JSON object in the following format:
    {{
        "request": "<YOUR-NEXT-INPUT>"
    }}
    """

    # Get the LLM's response
    response = model.generate_content(prompt, generation_config=config)
    llm_output = response.text

    try:
        # Attempt to parse the LLM's response as JSON
        llm_action = cleanup_json(str(llm_output))
        llm_action = json.loads(llm_action)
        return llm_action

    except json.JSONDecodeError:
        print(f"Error: Could not parse LLM response as JSON: {llm_output}")
        return None


def is_mission_done(mission_description, conversation_history):
    """Checks if the test mission has been completed.

    Args:
        mission_description (str): The description of the overall mission.
        conversation_history (list): A list of previous turns in the conversation.

    Returns:
        bool: True if the mission is complete, False otherwise.
    """

    # Construct the conversation history string for the prompt
    conversation_string = "\n".join(
        [
            f"{turn.get('role', 'unknown')}: {turn['content']}"
            for turn in conversation_history
        ]
    )

    prompt = f"""Mission description: {mission_description}

    Conversation history:
    {conversation_string}

    Based on the mission description and the conversation history, has the mission been successfully completed? 
    Answer only with "yes" or "no".
    """

    # Get the LLM's response
    response = model.generate_content(prompt, generation_config=config)
    llm_output = response.text.strip().lower()

    # Check if the LLM's response indicates mission completion
    if llm_output == "yes":
        return True
    elif llm_output == "no":
        return False
    else:
        print(f"Error: Unexpected LLM response: {llm_output}")
        return False


def evaluate_mission(
    mission_description, conversation_history, golden_response, golden_response_prompt
):
    """Evaluates the overall success of the test mission.

    Args:
        mission_description (str): The description of the overall mission.
        conversation_history (list): A list of previous turns in the conversation.
        golden_reponse: What should have been answered
        prompt: Golden response prompt

    Returns:
        dict: A dictionary containing the LLM's assessment of the mission.
    """

    current_time = datetime.now(tz=ZoneInfo("Europe/Berlin"))

    # Construct the conversation history string for the prompt
    conversation_string = "\n".join(
        [
            f"{turn.get('role', 'unknown')}: {turn['content']}"
            for turn in conversation_history
        ]
    )

    prompt = f"""
    
    {golden_response_prompt}

    You are evaluating the success of a test mission. The statement is the Mission description.

    Please provide an overall assessment of the mission's success: 
    - Was the mission completed successfully? 
    - What went well? 
    - What could have been improved?

    Today is {current_time.strftime('%A, %B %-d %Y')}. The current time is {current_time.strftime('%-H:%M')}.

    Statement (Mission description): {mission_description}
    Best-known response: {strip_references(golden_response)}

    Conversation history:
    {conversation_string}

    ALWAYS provide also following JSON object fields in your answer:
    {{
        "overall_success": "(Successful/Partially Successful/Failed)",
        "positive_observations": "List observations here",
        "areas_for_improvement": "List areas for improvement here"
    }}

    Comparison result:
    """

    # Get the LLM's response
    response = model.generate_content(prompt, generation_config=config)
    llm_output = response.text

    try:
        # Attempt to parse the LLM's assessment as JSON
        assessment = cleanup_json(str(llm_output))
        assessment = json.loads(assessment)
        return assessment

    except json.JSONDecodeError:
        print(f"Error: Could not parse LLM assessment as JSON: {llm_output}")
        return {"error": "Could not parse LLM assessment"}


print("Assessment module ready.")
