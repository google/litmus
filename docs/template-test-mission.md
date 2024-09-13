# Test Mission Template

The "Test Mission" template in Litmus provides a framework for evaluating your AI model's performance in multi-turn interactions, simulating a more realistic conversational flow or task-oriented dialogue. Unlike "Test Runs," which treat test cases as isolated units, "Test Missions" use the AI's responses from previous turns to guide subsequent interactions.

## Structure

A "Test Mission" template is structured similarly to a "Test Run" template but with key differences:

- **Template ID:** A unique identifier for your template.
- **Mission Data:** An array of missions, each containing:
  - **Mission:** A description of the overall task or goal the AI should achieve through the conversation.
  - **Mission Result:** The expected outcome or state at the end of a successful mission. This is used as a golden response to evaluate the entire conversation using an LLM.
  - **Filter** (optional): Comma-separated keywords or categories to organize and filter missions.
  - **Source** (optional): A source identifier for the mission.
  - **Block** (optional): A boolean (true/false) indicating whether this mission should be excluded from the run.
  - **Category** (optional): A category or label for the mission.
- **Mission Duration:** An integer specifying the maximum number of interaction turns allowed for each mission.
- **Request Payload:** A JSON object defining the structure of the API request sent to the AI model, including a placeholder `{query}` for the AI's generated input.
- **Pre-Request and Post-Request (optional):** JSON objects defining optional requests to be executed before and after each turn in the mission, respectively.
- **LLM Evaluation Prompt (optional):** A prompt guiding the LLM in evaluating the overall success of the mission based on the conversation history and the expected `Mission Result`. This prompt is used in the `evaluate_mission` function.

## Purpose

"Test Mission" templates aim to:

- **Assess conversational fluency and coherence:** Evaluate how well your AI maintains a natural and meaningful dialogue flow.
- **Test goal-oriented behavior:** Verify if your AI can achieve predefined tasks or goals through its interactions.
- **Identify limitations in context understanding:** Detect scenarios where your AI struggles to maintain context or follow instructions across multiple turns.
- **Evaluate task completion and outcome:** Assess the success of the mission based on the final state of the conversation and its alignment with the expected `Mission Result`.

## Scenarios

### Good Use Cases

- **Evaluating chatbot or dialogue system performance:** Assess how naturally and effectively your AI interacts in a conversational setting.
- **Testing task-oriented agents:** Verify if your AI can successfully complete tasks like booking appointments, ordering food, or providing customer support.
- **Evaluating multi-step reasoning and problem-solving abilities:** Assess your AI's capacity to break down complex goals into manageable steps and execute them through its interactions.

### Bad Use Cases

- **Single-turn interactions or simple question answering:** "Test Missions" are overkill for scenarios where each input is independent. Use "Test Runs" for these simpler cases.
- **Testing functionalities that don't involve conversational flow:** Avoid using "Test Missions" for tasks like sentiment analysis or text classification, which don't require multi-turn interactions.

## Starting a Test Mission

### API

1. Construct a JSON payload:
   ```json
   {
     "run_id": "your-unique-run-id",
     "template_id": "your-template-id",
     "pre_request": { ... }, // Optional
     "post_request": { ... }, // Optional
     "test_request": { ... },
     "template_type": "Test Mission",
     "mission_duration": <number_of_turns> // Required for Test Missions
   }
   ```
2. Send a POST request to the `/submit_run` endpoint.

### CLI (Advanced)

1. Gather the following information:
   - `RUN_ID`, `TEMPLATE_ID`, `GCP_PROJECT`, `GCP_REGION` (as with "Test Runs").
   - `MISSION_DURATION`: The maximum number of turns for each mission.
   - Optional: `PRE_REQUEST` and `POST_REQUEST` as JSON strings.
2. Invoke the Litmus worker using `gcloud`:
   ```bash
   gcloud run jobs execute litmus-worker \
       --project $GCP_PROJECT \
       --region $GCP_REGION \
       --set-env-vars RUN_ID=$RUN_ID,TEMPLATE_ID=$TEMPLATE_ID,TEMPLATE_TYPE="Test Mission",MISSION_DURATION=$MISSION_DURATION,PRE_REQUEST='$PRE_REQUEST',POST_REQUEST='$POST_REQUEST'
   ```

### CLI (Simple)

1. Get your `RUN_ID` and `TEMPLATE_ID`.
2. Run the following `litmus` CLI command:
   ```bash
   litmus start $TEMPLATE_ID $RUN_ID
   ```

### UI

1. Navigate to the "Start New Run" page.
2. Select your "Test Mission" template.
3. Enter your `RUN_ID`.
4. The UI should display input fields for `Mission Duration` and allow you to define or edit missions.
5. Submit the run.

## Configuration

- Modify the `Request Payload` to match your API.
- Define `Pre-Request` and `Post-Request` as needed.
- Craft the `LLM Evaluation Prompt` for assessing mission success.
- Adjust `Mission Duration` to control interaction length.

## Restarting and Deleting

Similar to "Test Runs," you can restart a "Test Mission" using the `/invoke_run` API endpoint or the UI's restart button. Deletion is done through the `/delete_run/<run_id>` endpoint or the UI's delete button.

## Additional Information

- Each mission's conversation history and the final LLM assessment are stored in Firestore, allowing for detailed analysis.
- Consider the ethical implications of multi-turn interactions when crafting your mission descriptions and LLM prompts.
- Experiment with different `Mission Durations` to find a balance between comprehensive evaluation and resource efficiency.
