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

# we use the version 3.10 of the python image 
FROM python:3.10

# Allow statements and log messages to immediately appear in the logs
ENV PYTHONUNBUFFERED True

# we define a work directory
WORKDIR /app

# we copy the code dir into the work directory
COPY requirements.txt /app

# we install the dependencies
RUN pip install --no-cache-dir -r requirements.txt

# we copy the code dir into the work directory
COPY . /app

# we execute the code with the following command
CMD [ "python", "main.py" ]