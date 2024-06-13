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
    <n-spin :show="show">
      <n-form ref="formRef" :model="formData" :rules="rules">
        <n-form-item label="Template ID" path="template_Id">
          <n-select
            v-model:value="formData.template_id"
            :options="templateOptions"
            @update:value="getTemplate"
          />
        </n-form-item>

        <n-form-item label="Run ID" path="run_Id">
          <n-input v-model:value="formData.run_id" placeholder="Please enter a run ID." />
        </n-form-item>

        <n-card>
          <div v-if="templateData.template_data.length > 0">
            <strong>Test Cases</strong>: {{ templateData.template_data.length }}
          </div>
          <n-tabs type="line" animated>
            <n-tab-pane name="Request Payload" tab="Request Payload">
              <!-- Request -->
              <json-editor-vue
                v-model="templateData.test_request"
                mode="text"
              ></json-editor-vue>
              The following tokens are available: {query} , {response} , {filter} ,
              {source} , {block} , {category}
            </n-tab-pane>
            <n-tab-pane name="Pre-Request (Optional)" tab="Pre-Request (Optional)">
              <!-- Test Pre-Request (Optional) -->
              <json-editor-vue
                v-model="templateData.test_pre_request"
                mode="text"
              ></json-editor-vue>
            </n-tab-pane>
            <n-tab-pane name="Post-Request (Optional)" tab="Post-Request (Optional)">
              <!-- Test Post-Request (Optional) -->
              <json-editor-vue
                v-model="templateData.test_post_request"
                mode="text"
              ></json-editor-vue>
            </n-tab-pane>
          </n-tabs>
        </n-card>

        <v-btn variant="flat" color="primary" class="mt-2" @click="submitForm"
          >Submit Run</v-btn
        >
      </n-form>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
import {
  NList,
  NListItem,
  NThing,
  NIcon,
  NSwitch,
  NCollapse,
  NCollapseItem,
  NCollapseTransition,
  NUpload,
  NTabs,
  NCard,
  NTabPane,
  NForm,
  NFormItem,
  NInput,
  NSelect,
  NButton,
  useMessage,
  NSpin,
} from "naive-ui";
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";
import JsonEditorVue from "json-editor-vue";

interface TemplateOption {
  label: string;
  value: string;
}

interface DataItem {
  query: string;
  response: string;
  filter: string;
  source: string;
  block: boolean;
  category: string;
}

interface PayloadItem {
  body: string;
  headers: {};
  method: string;
  url: string;
}

interface TemplateData {
  template_id: string;
  test_pre_request?: [];
  test_post_request?: [];
  test_request?: PayloadItem;
  template_data: DataItem[];
}

const templateData = ref<TemplateData>({
  template_id: "",
  template_data: [],
} as TemplateData);

const show = ref(false);
const router = useRouter();
const message = useMessage();

const formRef = ref();
const formData = ref({
  template_id: "",
  run_id: "",
  test_request: {},
  pre_request: {},
  post_request: {},
});
const rules = {
  template_id: {
    required: true,
    message: "Please select a template",
    trigger: "change",
  },
  run_id: {
    required: true,
    message: "Please enter a run ID",
    trigger: "blur",
  },
};

interface ValidationError {
  message: string;
  field?: string; // Optional field for specifying the error field
}

const templateOptions = ref<TemplateOption[]>([]);

const getTemplates = async () => {
  try {
    const response = await fetch("/templates");
    const data = await response.json();
    templateOptions.value = data.template_ids.map((id: string) => ({
      label: id,
      value: id,
    }));
  } catch (error) {
    console.error("Error fetching templates:", error);
    message.error("Failed to fetch templates");
  }
};

const getTemplate = async (value: string) => {
  try {
    const response = await fetch(`/templates/${value}`);
    if (!response.ok) {
      throw new Error("Template not found");
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

const submitForm = async () => {
  const form = formRef.value;
  show.value = true;
  form.validate(async (errors: ValidationError[]) => {
    if (!errors) {
      try {
        if (templateData.value.test_request) {
          formData.value.test_request = templateData.value.test_request;
        } else {
          // Handle the case where templateData.value.test_request is undefined
          // (e.g., assign an empty object or use a default value)
          formData.value.test_request = {};
        }
        if (templateData.value.test_pre_request) {
          formData.value.pre_request = templateData.value.test_pre_request;
        }
        if (templateData.value.test_post_request) {
          formData.value.post_request = templateData.value.test_post_request;
        }
        const response = await fetch("/submit_run", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify(formData.value),
        });
        if (response.ok) {
          message.success("Test run submitted successfully!");
          //router.push({ name: 'tests', params: { runId: formData.value.runId } });
          router.push({ name: "tests", params: {} });
        } else {
          const data = await response.json();
          message.error(data.error);
          show.value = false;
        }
      } catch (error) {
        message.error("Error submitting run.");
        show.value = false;
      }
    }
  });
};

onMounted(() => {
  getTemplates();
});
</script>
