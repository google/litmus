<!-- 
Copyright 2024 Google LLC

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License. 
-->

<template>
  <div>
    <!-- Loading spinner for the UI proxies table -->
    <n-spin :show="show">
      <n-table class="table-min-width" striped>
        <thead>
          <tr>
            <th>Name</th>
            <th>Created</th>
            <th>URI</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="proxy in proxies">
            <td>{{ proxy.name }}</td>
            <td>{{ proxy.created }}</td>
            <td>
              <strong>{{ proxy.uri }}</strong>
            </td>
            <td>
              <v-btn variant="flat" color="accent" class="mt-2" @click="showSetup(proxy.project_id, proxy.region, proxy.uri)">
                Set Up
              </v-btn>
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-spin>
    <n-divider />
    <!-- Link to proxy installation instructions -->
    Proxy installation instructions:
    <a href="https://google.github.io/litmus/proxy" target="_blank">https://google.github.io/litmus/proxy</a>

    <!-- Drawer for displaying setup instructions -->
    <n-drawer v-model:show="showDrawer" :width="980">
      <n-drawer-content :title="drawerTitle" :native-scrollbar="false" :width="996" closable>
        <div style="overflow: auto">
          <n-space vertical :size="16">
            <!-- Step 1: Install Vertex AI SDK -->
            1.) Install the Vertex AI SDK = Open a terminal window and enter the command below. You can also install it in a virtualenv
            <VCodeBlock
              code="pip install --upgrade google-cloud-aiplatform
gcloud auth application-default login
"
              highlightjs
              lang="bash"
              theme="stackoverflow-light"
            />

            <!-- Step 2: Use the provided code in your application -->
            2.) Use the following code in your application to request a model response
            <VCodeBlock :code="code" highlightjs lang="python" theme="stackoverflow-light" />
          </n-space>
        </div>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script lang="ts" setup>
import { NTable, NDivider, NSpace, NSpin, NDrawer, NDrawerContent } from 'naive-ui';
import { ref, onMounted } from 'vue';
import VCodeBlock from '@wdns/vue-code-block';

// Interface defining the structure of a proxy object
interface Proxy {
  project_id: string;
  region: string;
  uri: string;
  created: string;
  name: string;
}

// Reactive variable to store the list of proxies
const proxies = ref<Proxy[]>([]);

// Reactive variable controlling the visibility of the setup drawer
const showDrawer = ref(false);
// Reactive variable for the title of the setup drawer
const drawerTitle = ref('');
// Reactive variable controlling the visibility of the loading spinner for the table
const show = ref(false);
// Reactive variable containing the Python code snippet for the setup instructions
const code = ref(`import vertexai`);

/**
 * Function to display the setup instructions for a given proxy.
 *
 * @param {string} project - The project ID of the proxy.
 * @param {string} region - The region of the proxy.
 * @param {string} uri - The URI of the proxy.
 */
const showSetup = (project: string, region: string, uri: string) => {
  // Construct the Python code snippet with the provided project, region, and URI
  code.value =
    `import base64
import vertexai
from vertexai.generative_models import GenerativeModel, Part, SafetySetting, FinishReason
import vertexai.generative_models as generative_models
from vertexai.preview.prompts import Prompt
import uuid

def generate(CONTEXT):

  proxy_endpoint = '` +
    uri +
    `/litmus-context-{}'.format(CONTEXT)

  vertexai.init(project="` +
    project +
    `", location="` +
    region +
    `", api_endpoint=proxy_endpoint, api_transport="rest")

  prompt = Prompt(
    prompt_data=["""<YOUR PROMPT>"""],
    model_name="gemini-1.5-flash-001",
    generation_config=generation_config,
    safety_settings=safety_settings,
  )
  # Generate content using the assembled prompt. Change the index if you want
  # to use a different set in the variable value list.
  responses = prompt.generate_content(
      contents=prompt.assemble_contents(**prompt.variables[0]),
      stream=True,
  )

  for response in responses:
    print(response.text, end="")


generation_config = {
    "max_output_tokens": 8192,
    "temperature": 1,
    "top_p": 0.95,
}

safety_settings = [
    SafetySetting(
        category=SafetySetting.HarmCategory.HARM_CATEGORY_HATE_SPEECH,
        threshold=SafetySetting.HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE
    ),
    SafetySetting(
        category=SafetySetting.HarmCategory.HARM_CATEGORY_DANGEROUS_CONTENT,
        threshold=SafetySetting.HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE
    ),
    SafetySetting(
        category=SafetySetting.HarmCategory.HARM_CATEGORY_SEXUALLY_EXPLICIT,
        threshold=SafetySetting.HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE
    ),
    SafetySetting(
        category=SafetySetting.HarmCategory.HARM_CATEGORY_HARASSMENT,
        threshold=SafetySetting.HarmBlockThreshold.BLOCK_MEDIUM_AND_ABOVE
    ),
]

## Replace with your own CONTEXT from HTTP header ##
## For now we are generating a random CONTEXT     ##
CONTEXT = uuid.uuid4()

generate(CONTEXT)`;

  // Show the setup drawer and set the title
  showDrawer.value = true;
  drawerTitle.value = 'Set up instructions';
};

/**
 * Function to fetch the list of proxies from the backend API.
 */
const fetchProxies = () => {
  // Show the loading spinner
  show.value = true;

  // Fetch the proxies from the API
  fetch('/proxy/list_services')
    .then((response) => response.json())
    .then((data) => {
      // Update the proxies variable with the fetched data
      proxies.value = data;
      // Hide the loading spinner
      show.value = false;
    })
    .catch((error) => {
      console.error('Error fetching proxies:', error);
      // Hide the loading spinner
      show.value = false;
    });
};

// Fetch the list of proxies when the component is mounted
onMounted(() => {
  fetchProxies();
});
</script>

<style>
.pointer {
  cursor: pointer;
}

.n-code pre {
  background-color: #f0f0f0;
  padding-left: 1em;
}
</style>
