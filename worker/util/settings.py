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

import os

print("start settings")


class settings:

    # GCP Specific

    project_id = os.environ.get("GCP_PROJECT", "<INSERT-PROJECT>")
    location = os.environ.get("GCP_LOCATION", "us-central1")

    # AI Specific

    ai_location = os.environ.get("AI_LOCATION", "global")
    ai_default_model = os.environ.get("AI_DEFAULT_MODEL", "gemini-1.5-flash-002")
    ai_validation_model = os.environ.get("AI_DEFAULT_MODEL", "gemini-1.5-flash-002")
