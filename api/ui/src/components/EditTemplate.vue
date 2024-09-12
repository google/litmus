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
  <div class="add-update-template">
    <!-- Form to manage Litmus test templates -->
    <n-form ref="formRef" :model="templateData" :rules="rules" label-placement="top">
      <n-card>
        <n-space>
          <!-- Template ID Input -->
          <n-form-item label="Template ID" path="template_id">
            <n-input v-model:value="templateData.template_id" placeholder="Enter Template ID" :disabled="editMode" :allow-input="noSpace" />
          </n-form-item>

          <!-- Template Type Selection -->
          <n-form-item label="Type" path="template_type">
            <n-select
              v-model:value="templateData.template_type"
              :options="templateTypeOptions"
              @update:value="syncTabs"
              :disabled="editMode"
            />
          </n-form-item>
        </n-space>
      </n-card>

      <!-- Main Content Card: Test Cases, Request Payloads, LLM Prompts -->
      <n-card>
        <n-tabs type="line" ref="templateTabs" animated v-model:value="tabvalue">
          <!-- Test Cases Tab -->
          <n-tab-pane
            :name="templateData.template_type === 'Test Run' ? 'Test Cases' : 'Missions'"
            :tab="templateData.template_type === 'Test Run' ? 'Test Cases' : 'Missions'"
          >
            <!-- Mission Duration Input (only for Test Mission type) -->
            <n-form-item
              v-if="templateData.template_type === 'Test Mission'"
              label="Mission Duration (Number of loops)"
              path="mission_duration"
            >
              <n-input-number v-model:value="templateData.mission_duration" placeholder="Enter Mission Duration" />
            </n-form-item>
            <!-- Test Data / Mission Items -->
            <n-form-item>
              <n-collapse>
                <n-collapse-item v-for="(item, index) in templateData.template_data" :key="index" :title="item.query">
                  <div class="data-item">
                    <!-- Dynamic Label and Path for Query/Mission -->
                    <n-form-item
                      :label="templateData.template_type === 'Test Run' ? 'Query' : 'Mission'"
                      :path="`template_data.${index}.query`"
                    >
                      <n-input
                        v-model:value="item.query"
                        type="textarea"
                        :placeholder="templateData.template_type === 'Test Run' ? 'Enter Query' : 'Enter Mission'"
                      />
                    </n-form-item>

                    <!-- Dynamic Label and Path for Response/Mission Result -->
                    <n-form-item
                      :label="templateData.template_type === 'Test Run' ? 'Response' : 'Mission Result'"
                      :path="`template_data.${index}.response`"
                    >
                      <n-input
                        v-model:value="item.response"
                        type="textarea"
                        :placeholder="templateData.template_type === 'Test Run' ? 'Enter Response' : 'Enter Mission Result'"
                      />
                    </n-form-item>
                    <!-- Filter Input -->
                    <n-form-item label="Filter" :path="`template_data.${index}.filter`">
                      <n-input v-model:value="item.filter" placeholder="Enter Filter (comma-separated)" />
                    </n-form-item>
                    <!-- Source Input -->
                    <n-form-item label="Source" :path="`template_data.${index}.source`">
                      <n-input v-model:value="item.source" placeholder="Enter Source" />
                    </n-form-item>
                    <!-- Block Toggle Switch -->
                    <n-form-item label="Block" :path="`template_data.${index}.block`">
                      <n-switch v-model:value="item.block" />
                    </n-form-item>
                    <!-- Category Input -->
                    <n-form-item label="Category" :path="`template_data.${index}.category`">
                      <n-input v-model:value="item.category" placeholder="Enter Category" />
                    </n-form-item>
                    <!-- Delete Test Case Button -->
                    <v-btn variant="flat" color="accent" class="mt-2" @click="removeItem(index)"> Delete </v-btn>
                  </div>
                </n-collapse-item>
              </n-collapse>
            </n-form-item>

            <!-- Add Test Case and Upload JSON Buttons -->
            <n-space>
              <n-form-item>
                <v-btn variant="flat" color="secondary" class="mt-2" @click="addItem" v-if="templateData.template_type === 'Test Run'">
                  Add Test Case
                </v-btn>
                <v-btn variant="flat" color="secondary" class="mt-2" @click="addItem" v-if="templateData.template_type === 'Test Mission'">
                  Add Mission
                </v-btn>
              </n-form-item>
              <n-form-item>
                <n-upload @before-upload="handleFileUpload">
                  <v-btn variant="flat" color="secondary" class="mt-2"> Upload JSON </v-btn>
                </n-upload>
              </n-form-item>
            </n-space>
            <!-- Link to JSON Template -->
            <a href="https://storage.googleapis.com/litmus-cloud/assets/template.json" target="_blank"> Click here for JSON Template </a>
          </n-tab-pane>

          <!-- Request Payload Tab -->
          <n-tab-pane name="Request Payload" tab="Request Payload">
            <!-- Test Request Button -->
            <n-space justify="end" size="large">
              <v-btn variant="flat" color="secondary" class="mt-2" @click="testRequest"> Test Request </v-btn>
            </n-space>
            <n-divider />
            <!-- JSON Editor for Request Payload -->
            <json-editor-vue v-model="templateData.test_request" mode="text"></json-editor-vue>
            <!-- Available Tokens Information -->
            The following tokens are available: {query} , {response} , {filter} , {source} , {block} , {category}, {auth_token}
          </n-tab-pane>

          <!-- Pre-Request (Optional) Tab -->
          <n-tab-pane name="Pre-Request (Optional)" tab="Pre-Request (Optional)">
            <!-- JSON Editor for Pre-Request Payload -->
            <json-editor-vue v-model="templateData.test_pre_request" mode="text"></json-editor-vue>
          </n-tab-pane>

          <!-- Post-Request (Optional) Tab -->
          <n-tab-pane name="Post-Request (Optional)" tab="Post-Request (Optional)">
            <!-- JSON Editor for Post-Request Payload -->
            <json-editor-vue v-model="templateData.test_post_request" mode="text"></json-editor-vue>
          </n-tab-pane>

          <!-- LLM Evaluation Prompt (Optional) Tab -->
          <n-tab-pane name="LLM Evaluation Prompt" tab="LLM Evaluation Prompt">
            <!-- Textarea for LLM Evaluation Prompt -->
            <n-input
              v-model:value="templateData.template_llm_prompt"
              type="textarea"
              :autosize="{
                minRows: 3
              }"
            />
          </n-tab-pane>
        </n-tabs>
      </n-card>

      <!-- Input and Output Field Selection Card -->
      <n-card>
        <n-space justify="space-around" size="large">
          <!-- Input Field Selection -->
          <strong>Input Field</strong>
          <n-button @click="showInputUI = true" dashed>
            {{ templateData.template_input_field }}
          </n-button>
          <n-drawer v-model:show="showInputUI" :width="500" placement="left">
            <n-drawer-content title="Input" closable>
              <JsonTreeView :json="templateData.test_request" :maxDepth="10" @selected="onSelectedInput" />
            </n-drawer-content>
          </n-drawer>

          <!-- Output Field Selection -->
          <strong>Output Field</strong>
          <n-button @click="testRequest" dashed>
            {{ templateData.template_output_field }}
          </n-button>
          <n-drawer v-model:show="showOutputUI" :width="500" placement="left">
            <n-drawer-content title="Output" closable>
              <JsonTreeView :json="test_response" :maxDepth="10" @selected="onSelectedOutput" />
            </n-drawer-content>
          </n-drawer>
        </n-space>
      </n-card>

      <!-- Submit/Update Template Button -->
      <n-form-item>
        <v-btn variant="flat" color="primary" class="mt-2" @click="submitForm"> {{ editMode ? 'Update' : 'Add' }} Template </v-btn>
      </n-form-item>
    </n-form>
  </div>
</template>

<script lang="ts" setup>
import {
  NSwitch,
  NCollapse,
  NCollapseItem,
  NUpload,
  NForm,
  NFormItem,
  NInput,
  NButton,
  NTabs,
  NCard,
  NTabPane,
  NDrawer,
  NDrawerContent,
  NSpace,
  NDivider,
  useMessage,
  NSelect,
  NInputNumber
} from 'naive-ui';
import type { TabsInst } from 'naive-ui';
import type { UploadFileInfo } from 'naive-ui';
import { ref, onMounted, onUnmounted, nextTick } from 'vue';
import JsonEditorVue from 'json-editor-vue';
import { JsonTreeView } from 'json-tree-view-vue3';
import 'json-tree-view-vue3/dist/style.css';

const templateTabs = ref<TabsInst | null>(null);
const tabvalue = ref();

// Syncing correct Tabs
const syncTabs = (value: string) => {
  if (value == 'Test Mission') {
    tabvalue.value = 'Missions';
  } else {
    tabvalue.value = 'Test Cases';
  }
  nextTick(() => templateTabs.value?.syncBarPosition());
};

// Interface for Test Data Items
interface DataItem {
  query: string;
  response: string;
  filter: string;
  source: string;
  block: boolean;
  category: string;
}

// Type for Primitive Data Types
type PrimitiveTypes = string | number | boolean | null;

// Interface for Selected Data in JSON Tree View
interface SelectedData {
  key: string;
  value: PrimitiveTypes;
  path: string;
}

// Interface for Template Data Structure
interface TemplateData {
  template_id: string;
  test_pre_request?: [];
  test_post_request?: [];
  test_request?: string;
  template_data: DataItem[];
  template_input_field: string;
  template_output_field: string;
  template_llm_prompt: string;
  template_type: string; // Added template type
  mission_duration?: number; // Added mission duration, optional
}

// Emits 'close' event to parent component
const emit = defineEmits(['close']);

// Refs for Form, Loading State, UI Elements
const formRef = ref();
const loading = ref(false);
const showInputUI = ref(false);
const showOutputUI = ref(false);

// Naive UI Message Instance
const message = useMessage();

// Test Response Data
let test_response: string = '';

// Template Type Options for Dropdown
const templateTypeOptions = [
  {
    label: 'Test Run',
    value: 'Test Run'
  },
  {
    label: 'Test Mission',
    value: 'Test Mission'
  }
];

// Template Data Object with Default Values
const templateData = ref<TemplateData>({
  template_id: '',
  template_data: [],
  test_request: '',
  template_input_field: '',
  template_output_field: '',
  template_llm_prompt: '',
  template_type: 'Test Run', // Default template type
  mission_duration: undefined // Mission duration is optional
});

// Edit Mode Flag
const editMode = ref(false);

// Form Validation Rules
const rules = {
  template_id: {
    required: true,
    message: 'Please enter a Template ID',
    trigger: ['blur', 'input']
  },
  mission_duration: {
    // Validation rule for mission_duration
    required: true,
    type: 'number',
    message: 'Please enter Mission Duration',
    trigger: ['blur', 'input']
  }
  // ... add validation rules for other fields
};

/**
 * Handles file uploads, validating for JSON format and structure.
 * @param data - Upload event data including file and fileList.
 */
const handleFileUpload = (data: { file: UploadFileInfo; fileList: UploadFileInfo[] }) => {
  const fileData = data.file;
  const fileType = fileData.type;

  if (fileType === 'application/json') {
    if (fileData.file) {
      const file = fileData.file;
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const jsonData = JSON.parse(e.target?.result as string);
          if (validateJsonStructure(jsonData)) {
            templateData.value.template_data = jsonData;
          } else {
            console.error('Invalid JSON structure');
          }
        } catch (error) {
          console.error('Error parsing JSON:', error);
        }
      };
      reader.readAsText(file);
    } else {
      console.warn('No file selected');
    }
  } else {
    console.warn('Uploaded file is not a JSON file');
  }
};

/**
 * Validates the structure of the uploaded JSON data.
 * @param data - The parsed JSON data.
 * @returns True if the structure is valid, otherwise false.
 */
const validateJsonStructure = (data: any[]) => {
  if (!Array.isArray(data)) {
    return false;
  }
  return data.every((item) => {
    return (
      typeof item.query === 'string' &&
      typeof item.response === 'string' &&
      Array.isArray(item.filter) &&
      typeof item.source === 'string' &&
      typeof item.block === 'string' &&
      typeof item.category === 'string'
    );
  });
};

/**
 * Fetches a template by ID from the API.
 * @param templateId - The ID of the template to fetch.
 */
const getTemplate = async (templateId: string) => {
  try {
    const response = await fetch(`/templates/${templateId}`);
    if (!response.ok) {
      throw new Error('Template not found');
    }
    const data = await response.json();
    templateData.value = data;
    if (props.templateId) {
      templateData.value.template_id = props.templateId;
    }
    if (!data.template_data) {
      templateData.value.template_data = [];
    }
    editMode.value = true;
    syncTabs(templateData.value.template_type);
  } catch (error) {
    throw new Error('Failed to get Templates');
  }
};

/**
 * Submits the form data to the API, either creating or updating a template.
 */
const submitForm = async () => {
  loading.value = true;

  try {
    const dataToSend = {
      ...templateData.value,
      template_data: templateData.value.template_data.map((item) => ({
        ...item,
        filter: typeof item.filter === 'string' ? item.filter.split(',') : []
      }))
    };

    const response = await fetch(editMode.value ? '/update_template' : '/add_template', {
      method: editMode.value ? 'PUT' : 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(dataToSend)
    });

    if (!response.ok) {
      throw new Error('Failed to submit form');
    }

    const responseData = await response.json();
    console.log('Form submitted successfully:', responseData);
  } catch (error) {
    console.error('Error submitting form:', error);
  } finally {
    loading.value = false;
    emit('close');
  }
};

/**
 * Adds a new empty test case item to the template data.
 */
const addItem = () => {
  if (templateData.value.template_type == 'Test Run') {
    templateData.value.template_data.push({
      query: 'Enter your query',
      response: '',
      filter: '',
      source: '',
      block: false,
      category: ''
    });
  } else {
    templateData.value.template_data.push({
      query: 'Enter your mission',
      response: '',
      filter: '',
      source: '',
      block: false,
      category: ''
    });
  }
};

/**
 * Sends a test request based on the provided payload in the template data.
 */
const testRequest = async () => {
  const reqString = templateData.value.test_request;

  if (typeof reqString === 'string') {
    try {
      const req = JSON.parse(reqString);

      const response = await fetch(req.url, {
        method: req.method,
        headers: req.headers,
        body: JSON.stringify(req.body)
      });

      if (response.ok) {
        const resptemp = await response.json();
        test_response = JSON.stringify(resptemp);
        showOutputUI.value = true;
      } else {
        message.error('Error: ' + response.status + ' ' + response.statusText);
      }
    } catch (parseError) {
      console.error('Error parsing JSON:', parseError);
      message.error('Invalid JSON format in Request Payload');
    }
  } else {
    console.error('test_request is undefined');
    message.error('Request Payload is empty');
  }
};

/**
 * Handles the selection of an input field from the JSON Tree View.
 * @param event - Selection event data.
 */
const onSelectedInput = (event: SelectedData) => {
  showInputUI.value = false;
  templateData.value.template_input_field = event.path.substring(2);
};

/**
 * Handles the selection of an output field from the JSON Tree View.
 * @param event - Selection event data.
 */
const onSelectedOutput = (event: SelectedData) => {
  showOutputUI.value = false;
  templateData.value.template_output_field = event.path.substring(2);
};

// Props Definition
const props = defineProps({
  templateId: {
    type: String,
    required: false
  }
});

/**
 * Removes a test case item from the template data by index.
 * @param index - The index of the item to remove.
 */
const removeItem = (index: number) => {
  templateData.value.template_data.splice(index, 1);
};

// Lifecycle Hook: onMounted
onMounted(() => {
  if (props.templateId) {
    getTemplate(props.templateId);
  } else {
    templateData.value = {
      template_id: '',
      template_data: [],
      test_request: JSON.stringify({
        body: { query: '{query}' },
        headers: {
          'Content-Type': 'application/json'
        },
        method: 'POST',
        url: 'https://example.com/request' // Placeholder URL
      }),
      template_llm_prompt: `You are a thorough quality inspector. Your task is to compare a statement about some topic to a golden response. The statement and the response can have different formats. Both statement and response are in German. You inspect the statement and the response to find out:
- has the question been answered at all?
- does the statement contradict the response?
- is the statement content-wise equivalent to the response, even it might have additional information?
- does the statement have additional information not contained in the response?
- is the statement missing information that is contained in the response?
- to what degree is the structure and wording of the statement similar to the response, even if content may be different?
Statements that are similar are estimated closer to 1 and statements that have different structure or wording are estimated closer to 0.
You MUST provide your output in JSON format. Do not provide any additional output.
This is what the JSON should look like:
{
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
}


Here is an example:
Statement: The soccer player B was Footballer of the Year in 2011 and 2012.
Best-known response: B was the Soccer Player of the Year in 2010 and 2012.

Comparison result:
{
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
}


Here is another example:
Statement: Police in canton A were called to a private residence yesterday for a disturbance.
Best-known response: Police in canton A responded to a disturbance yesterday.

Comparison result:
{
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
}


Here is another example:
Statement: C has been CEO of D since 2010.
Best-known response: C was appointed CEO of D in 2010.

Comparison result:
{
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
}

Here is another example:
Statement: I cannot answer this question.
Best-known response: The city of X was founded in 1833.

Comparison result:
{
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
}`,
      template_input_field: 'INCOMPLETE',
      template_output_field: 'INCOMPLETE',
      template_type: 'Test Run', // Default template type
      mission_duration: undefined // Mission duration is optional
    };
  }
});

const noSpace = (value: string) => !/ /g.test(value);

// Lifecycle Hook: onUnmounted
onUnmounted(() => {
  // ... Perform any cleanup if necessary
});
</script>
