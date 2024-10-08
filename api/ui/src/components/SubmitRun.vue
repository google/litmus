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
          <n-input v-model:value="formData.run_id" placeholder="Please enter a run ID." :allow-input="noSpace" />
        </n-form-item>

        <!-- Template details and test request card -->
        <n-card v-if="templateData.template_data.length > 0">
          <div>
            <strong>{{ templateData.template_type }}s:</strong> {{ templateData.template_data.length }} | <strong>Input Field:</strong>
            {{ templateData.template_input_field }} | <strong>Output Field:</strong> {{ templateData.template_output_field }}
            <template v-if="templateData.template_type === 'Test Mission'"
              >| <strong>Duration (loops):</strong> {{ templateData.mission_duration }}</template
            >
          </div>

          <!-- Tabs for Request Payload, Pre-Request, and Post-Request, Evaluation Types -->
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

            <!-- Evaluation Types Tab -->
            <n-tab-pane name="Evaluation" tab="Evaluation">
              <!-- Evaluation Types Checkboxes -->
              <h3>Evaluation Types</h3>
              <n-checkbox
                @update:checked="updateEvaluationType('llm_assessment', $event)"
                :checked="formData.evaluation_types.llm_assessment"
              >
                Custom LLM Evaluation
              </n-checkbox>
              <n-checkbox @update:checked="updateEvaluationType('ragas', $event)" :checked="formData.evaluation_types.ragas">
                Ragas
              </n-checkbox>
              <n-checkbox @update:checked="toggleDeepEvalOptions($event)" :checked="showDeepEvalOptions"> DeepEval </n-checkbox>

              <div v-if="showDeepEvalOptions">
                <h4>DeepEval Metrics</h4>
                <n-checkbox
                  v-for="metric in deepevalMetrics"
                  :key="metric"
                  :checked="formData.evaluation_types.deepeval.includes(metric)"
                  @update:checked="updateDeepEvalMetric(metric, $event)"
                >
                  {{ metric }}
                </n-checkbox>
              </div>
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
import { NTabs, NCard, NTabPane, NForm, NFormItem, NInput, NSelect, useMessage, NSpin, NCheckbox } from 'naive-ui';
import { ref, onMounted, watch } from 'vue';
import { useRouter } from 'vue-router';
import JsonEditorVue from 'json-editor-vue';

// Interface for template options in the dropdown
interface TemplateOption {
  label: string; // Label of the option
  value: string; // Value of the option
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
  body: string; // Body of the request/response
  headers: {}; // Headers of the request/response
  method: string; // HTTP method (e.g., GET, POST)
  url: string; // URL for the request
}

// Interface for template data
interface TemplateData {
  template_id: string; // ID of the template
  test_pre_request?: []; // Pre-request data (optional)
  test_post_request?: []; // Post-request data (optional)
  test_request?: PayloadItem; // Test request data
  template_data: DataItem[]; // Array of test data items
  template_input_field: string; // Input field for the template
  template_output_field: string; // Output field for the template
  template_llm_prompt: string; // LLM prompt for the template
  template_type: string; // Template type (e.g., 'Test Run', 'Test Mission')
  mission_duration?: number; // Mission duration (optional, for 'Test Mission' type)
  evaluation_types: {
    llm_assessment: boolean;
    ragas: boolean;
    deepeval: string[];
  };
}

// Reactive variable for storing template data
const templateData = ref<TemplateData>({
  template_id: '',
  template_data: [],
  template_input_field: '',
  template_output_field: '',
  template_llm_prompt: '',
  template_type: '',
  mission_duration: undefined,
  evaluation_types: {
    llm_assessment: false,
    ragas: false,
    deepeval: []
  }
} as TemplateData);

// Reactive variable to control loading spinner visibility
const show = ref(false);

// Router instance for navigation
const router = useRouter();

// Message instance for displaying notifications
const message = useMessage();

// Reference to the form element
const formRef = ref();

// Define the type for evaluation_types
type EvaluationTypes = {
  llm_assessment: boolean;
  ragas: boolean;
  deepeval: string[];
};

// Reactive variable for storing form data
const formData = ref({
  template_id: '',
  run_id: '',
  test_request: {},
  pre_request: {},
  post_request: {},
  evaluation_types: {
    llm_assessment: false,
    ragas: false,
    deepeval: [] as string[]
  } as EvaluationTypes // Apply the EvaluationTypes type here
});

// DeepEval Metrics
const deepevalMetrics = [
  'answer_relevancy',
  'faithfulness',
  'contextual_precision',
  'contextual_recall',
  'contextual_relevancy',
  'hallucination',
  'bias',
  'toxicity'
];

// Control visibility of DeepEval metric options
const showDeepEvalOptions = ref(false);

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
  message: string; // Validation error message
  field?: string; // Optional field name where the error occurred
}

// Reactive variable for storing template options for the dropdown
const templateOptions = ref<TemplateOption[]>([]);

// Function to fetch available templates from the API
const getTemplates = async () => {
  try {
    const response = await fetch('/templates/');
    const data = await response.json();
    // Extract template IDs from the 'templates' array
    templateOptions.value = data.templates.map((template: { template_id: string; template_type: string }) => ({
      label: template.template_type + ' : ' + template.template_id,
      value: template.template_id
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

    // Update formData.evaluation_types with values from the template
    formData.value.evaluation_types = { ...templateData.value.evaluation_types };

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
        // Check if test_request exists in templateData, otherwise assign an empty object to avoid errors
        formData.value.test_request = templateData.value.test_request || {};
        if (templateData.value.test_pre_request) {
          formData.value.pre_request = templateData.value.test_pre_request;
        }
        if (templateData.value.test_post_request) {
          formData.value.post_request = templateData.value.test_post_request;
        }

        // Send the test run data to the API
        const response = await fetch('/runs/submit', {
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

// Function to update evaluation types in formData
const updateEvaluationType = (type: keyof EvaluationTypes, checked: boolean) => {
  if (type === 'deepeval') {
    // Handle deepeval (string array) case
    if (checked) {
      // If checked, add all deepevalMetrics to the array
      formData.value.evaluation_types.deepeval = [...deepevalMetrics];
    } else {
      // If unchecked, clear the deepeval array
      formData.value.evaluation_types.deepeval = [];
    }
  } else {
    // Handle llm_assessment and ragas (boolean) cases
    formData.value.evaluation_types[type] = checked;
  }
};

// Function to toggle DeepEval options visibility
const toggleDeepEvalOptions = (checked: boolean) => {
  showDeepEvalOptions.value = checked;
};

// Function to update DeepEval metrics in formData
const updateDeepEvalMetric = (metric: string, checked: boolean) => {
  if (checked) {
    formData.value.evaluation_types.deepeval.push(metric);
  } else {
    const index = formData.value.evaluation_types.deepeval.indexOf(metric);
    if (index > -1) {
      formData.value.evaluation_types.deepeval.splice(index, 1);
    }
  }
};

const noSpace = (value: string) => !/ /g.test(value);

// Fetch available templates when the component is mounted
onMounted(() => {
  getTemplates();
});

// Watch for changes in showDeepEvalOptions
watch(showDeepEvalOptions, (newValue) => {
  // If showDeepEvalOptions is false, clear the selected DeepEval metrics
  if (!newValue) {
    formData.value.evaluation_types.deepeval = [];
  }
});
</script>
