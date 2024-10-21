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

"""Centralized application settings for the Litmus API testing tool."""

import os


class Settings:
    """
    A class to store and manage configuration settings for the Litmus application.

    Settings are loaded from environment variables or fall back to default values.
    """

    # GCP Specific
    project_id: str = os.environ.get("GCP_PROJECT", "<INSERT-PROJECT>")
    """GCP Project ID. Defaults to "<INSERT-PROJECT>"."""
    region: str = os.environ.get("GCP_REGION", "us-central1")
    """GCP Region. Defaults to "us-central1"."""

    # AI Specific
    ai_location: str = os.environ.get("AI_LOCATION", "global")
    """Location for AI models. Defaults to "global"."""
    ai_default_model: str = os.environ.get("AI_DEFAULT_MODEL", "gemini-1.5-flash-002")
    """Default AI Model. Defaults to "gemini-1.5-flash-002"."""
    ai_validation_model: str = os.environ.get(
        "AI_VALIDATION_MODEL", "gemini-1.5-flash-002"
    )
    """AI Model for validation of responses. Defaults to "gemini-1.5-flash-002"."""

    # Application general
    disable_auth: bool = os.getenv("DISABLE_AUTH", "False") == "True"
    """Flag to enable/disable authentication. 
    Defaults to False (authentication enabled)
    """
    auth_user: str = os.environ.get("USERNAME", "admin")
    """Username for authentication. Defaults to "admin"."""
    auth_pass: str = os.environ.get("PASSWORD", "admin")
    """Password for authentication. Defaults to "admin"."""


# Create an instance of the Settings class to access settings
settings = Settings()
