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
    <!-- Loading spinner while submitting the form -->
    <n-spin :show="show">
      <!-- Form for submitting a new test run -->
      <n-form ref="formRef" :model="formData" :rules="rules">
        <!-- Template ID selection -->
        <n-form-item label="Template ID" path="template_Id">
          <n-select v-model:value="formData.template_id" :options="templateOptions" @update:value="getTemplate" />
        </n-form-item>

        <!-- Run ID input -->
        <n-form-item label="Run ID" path="run_Id">
          <n-input v-model:value="formData.run_id" placeholder="Please enter a run ID." />
        </n-form-item>

        <!-- Template details and test request card -->
        <n-card>
          <!-- Display template details if available -->
          <div v-if="templateData.template_data.length > 0">
            <strong>Test Cases</strong>: {{ templateData.template_data.length }} |
            <strong> Input Field:</strong>
            {{ templateData.template_input_field }} | <strong>Output Field:</strong>
            {{ templateData.template_output_field }}
          </div>

          <!-- Tabs for Request Payload, Pre-Request, and Post-Request -->
          <n-tabs type="line" animated>
            <!-- Request Payload tab -->
            <n-tab-pane name="Request Payload" tab="Request Payload">
              <!-- JSON editor for editing the request payload -->
              <json-editor-vue v-model="templateData.test_request" mode="text"></json-editor-vue>
              <!-- Available tokens for the request payload -->
              The following tokens are available: {query} , {response} , {filter} , {source} , {block} , {category}
            </n-tab-pane>

            <!-- Pre-Request tab (optional) -->
            <n-tab-pane name="Pre-Request (Optional)" tab="Pre-Request (Optional)">
              <!-- JSON editor for editing the pre-request payload (optional) -->
              <json-editor-vue v-model="templateData.test_pre_request" mode="text"></json-editor-vue>
            </n-tab-pane>

            <!-- Post-Request tab (optional) -->
            <n-tab-pane name="Post-Request (Optional)" tab="Post-Request (Optional)">
              <!-- JSON editor for editing the post-request payload (optional) -->
              <json-editor-vue v-model="templateData.test_post_request" mode="text"></json-editor-vue>
            </n-tab-pane>
          </n-tabs>
        </n-card>

        <!-- Submit button to initiate the test run -->
        <v-btn variant="flat" color="primary" class="mt-2" @click="submitForm"> Submit Run </v-btn>
      </n-form>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
import { NTabs, NCard, NTabPane, NForm, NFormItem, NInput, NSelect, useMessage, NSpin } from 'naive-ui';
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';
import JsonEditorVue from 'json-editor-vue';

// Interface for template options in the dropdown
interface TemplateOption {
  label: string;
  value: string;
}

// Interface for data items within the template
interface DataItem {
  query: string;
  response: string;
  filter: string;
  source: string;
  block: boolean;
  category: string;
}

// Interface for request/response payload items
interface PayloadItem {
  body: string;
  headers: {};
  method: string;
  url: string;
}

// Interface for template data
interface TemplateData {
  template_id: string;
  test_pre_request?: [];
  test_post_request?: [];
  test_request?: PayloadItem;
  template_data: DataItem[];
  template_input_field: string;
  template_output_field: string;
  template_llm_prompt: string;
}

// Reactive variable for storing template data
const templateData = ref<TemplateData>({
  template_id: '',
  template_data: [],
  template_input_field: '',
  template_output_field: '',
  template_llm_prompt: ''
} as TemplateData);

// Reactive variable to control loading spinner visibility
const show = ref(false);

// Router instance for navigation
const router = useRouter();

// Message instance for displaying notifications
const message = useMessage();

// Reference to the form element
const formRef = ref();

// Reactive variable for storing form data
const formData = ref({
  template_id: '',
  run_id: '',
  test_request: {},
  pre_request: {},
  post_request: {},
  template_input_field: '',
  template_output_field: '',
  template_llm_prompt: ''
});

// Form validation rules
const rules = {
  template_id: {
    required: true,
    message: 'Please select a template',
    trigger: 'change'
  },
  run_id: {
    required: true,
    message: 'Please enter a run ID',
    trigger: 'blur'
  }
};

// Interface for validation errors
interface ValidationError {
  message: string;
  field?: string; // Optional field for specifying the error field
}

// Reactive variable for storing template options for the dropdown
const templateOptions = ref<TemplateOption[]>([]);

// Function to fetch available templates from the API
const getTemplates = async () => {
  try {
    const response = await fetch('/templates');
    const data = await response.json();
    templateOptions.value = data.template_ids.map((id: string) => ({
      label: id,
      value: id
    }));
  } catch (error) {
    console.error('Error fetching templates:', error);
    message.error('Failed to fetch templates');
  }
};

// Function to fetch template details based on the selected template ID
const getTemplate = async (value: string) => {
  try {
    const response = await fetch(`/templates/${value}`);
    if (!response.ok) {
      throw new Error('Template not found');
    }
    const data = await response.json();
    templateData.value = data;
    if (!data.template_data) {
      templateData.value.template_data = [];
    }
  } catch (error) {
    // Handle error (e.g., display error message, redirect)
  }
};

// Function to submit the test run form
const submitForm = async () => {
  const form = formRef.value;
  show.value = true;
  form.validate(async (errors: ValidationError[]) => {
    if (!errors) {
      try {
        // Populate form data with template details
        formData.value.template_input_field = templateData.value.template_input_field;
        formData.value.template_output_field = templateData.value.template_output_field;
        formData.value.template_llm_prompt = templateData.value.template_llm_prompt;
        // Check if test_request exists in templateData, otherwise assign an empty object to avoid errors
        formData.value.test_request = templateData.value.test_request || {};
        if (templateData.value.test_pre_request) {
          formData.value.pre_request = templateData.value.test_pre_request;
        }
        if (templateData.value.test_post_request) {
          formData.value.post_request = templateData.value.test_post_request;
        }
        // Send the test run data to the API
        const response = await fetch('/submit_run', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(formData.value)
        });
        if (response.ok) {
          message.success('Test run submitted successfully!');
          // Navigate to the tests page after successful submission
          router.push({ name: 'tests', params: {} });
        } else {
          const data = await response.json();
          message.error(data.error);
          show.value = false;
        }
      } catch (error) {
        message.error('Error submitting run.');
        show.value = false;
      }
    }
  });
};

// Fetch available templates when the component is mounted
onMounted(() => {
  getTemplates();
});
</script>
