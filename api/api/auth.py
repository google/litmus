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

"""This module handles authentication."""

from flask_httpauth import HTTPBasicAuth
from werkzeug.security import generate_password_hash, check_password_hash
from util.settings import settings

auth = HTTPBasicAuth()

# Authentication setup
if not settings.disable_auth:
    users = {settings.auth_user: generate_password_hash(settings.auth_pass)}


# Authentication verification callback function
@auth.verify_password
def verify_password(username, password):
    """Verifies user password for basic authentication.

    Args:
        username: Username provided in the request.
        password: Password provided in the request.

    Returns:
        Username if authentication is successful, otherwise None.
    """
    if not settings.disable_auth:
        return (
            username
            if username in users and check_password_hash(users.get(username), password)
            else None
        )
    return username  # Disable authentication


def login_required(x):
    return auth.login_required(x)
