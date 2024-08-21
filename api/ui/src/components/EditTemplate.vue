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
    <n-form ref="formRef" :model="templateData" :rules="rules" label-placement="top">
      <n-form-item label="Template ID" path="template_id">
        <n-input
          v-model:value="templateData.template_id"
          placeholder="Enter Template ID"
          :disabled="editMode"
        />
      </n-form-item>
      <n-card>
        <n-tabs type="line" animated>
          <n-tab-pane name="Test Cases" tab="Test Cases">
            <!-- Test Requests -->
            <!-- Data Items -->
            <n-form-item>
              <n-collapse>
                <n-collapse-item
                  v-for="(item, index) in templateData.template_data"
                  :key="index"
                  :title="item.query"
                >
                  <div class="data-item">
                    <n-form-item label="Query" :path="`template_data.${index}.query`">
                      <n-input v-model:value="item.query" placeholder="Enter Query" />
                    </n-form-item>
                    <n-form-item
                      label="Response"
                      :path="`template_data.${index}.response`"
                    >
                      <n-input
                        v-model:value="item.response"
                        type="textarea"
                        placeholder="Enter Response"
                      />
                    </n-form-item>
                    <n-form-item label="Filter" :path="`template_data.${index}.filter`">
                      <n-input
                        v-model:value="item.filter"
                        placeholder="Enter Filter (comma-separated)"
                      />
                    </n-form-item>
                    <n-form-item label="Source" :path="`template_data.${index}.source`">
                      <n-input v-model:value="item.source" placeholder="Enter Source" />
                    </n-form-item>
                    <n-form-item label="Block" :path="`template_data.${index}.block`">
                      <n-switch v-model:value="item.block" />
                    </n-form-item>
                    <n-form-item
                      label="Category"
                      :path="`template_data.${index}.category`"
                    >
                      <n-input
                        v-model:value="item.category"
                        placeholder="Enter Category"
                      />
                    </n-form-item>
                    <v-btn
                      variant="flat"
                      color="accent"
                      class="mt-2"
                      @click="removeItem(index)"
                      >Delete</v-btn
                    >
                  </div>
                </n-collapse-item>
              </n-collapse>
            </n-form-item>
            <n-form-item>
              <v-btn variant="flat" color="secondary" class="mt-2" @click="addItem"
                >Add Test Case</v-btn
              >
            </n-form-item>
            <n-form-item>
              <n-upload @before-upload="handleFileUpload">
                <v-btn variant="flat" color="secondary" class="mt-2">Upload JSON</v-btn>
              </n-upload>
            </n-form-item>
            <a href="https://storage.googleapis.com/litmus-cloud/assets/template.json" target="_blank">Click here for JSON Template</a>
          </n-tab-pane>
          <n-tab-pane name="Request Payload" tab="Request Payload">
            <!-- Request -->
            <json-editor-vue
              v-model="templateData.test_request"
              mode="text"
            ></json-editor-vue>
            The following tokens are available: {query} , {response} , {filter} , {source}
            , {block} , {category}
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
      <n-form-item>
        <v-btn variant="flat" color="primary" class="mt-2" @click="submitForm">
          {{ editMode ? "Update" : "Add" }} Template</v-btn
        >
      </n-form-item>
    </n-form>
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
  NForm,
  NFormItem,
  NInput,
  NButton,
  NTabs,
  NCard,
  NTabPane,
} from "naive-ui";
import type { UploadFileInfo } from "naive-ui";
import { ref, onMounted, onUnmounted, watch } from "vue";
import { useRoute, useRouter } from "vue-router";
import JsonEditorVue from "json-editor-vue";

interface DataItem {
  query: string;
  response: string;
  filter: string;
  source: string;
  block: boolean;
  category: string;
}

interface PayloadItem {
  body: {};
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

const emit = defineEmits(["close"]);
const route = useRoute();
const router = useRouter();
const formRef = ref();
const loading = ref(false);
const templateData = ref<TemplateData>({
  template_id: "",
  template_data: [],
} as TemplateData);
const editMode = ref(false);
const templateIdInput = ref<HTMLElement | null>(null);

const rules = {
  template_id: {
    required: true,
    message: "Please enter a Template ID",
    trigger: ["blur", "input"],
  },
  // ... add validation rules for other fields
};

const handleFileUpload = (data: { file: UploadFileInfo; fileList: UploadFileInfo[] }) => {
  // Access the uploaded file information
  const fileData = data.file;
  const fileName = fileData.name;
  const fileType = fileData.type;

  // Check if the file is a JSON file
  if (fileType === "application/json") {
    if (fileData.file) {
      const file = fileData.file; // Assuming fileData.file contains the File object
      const reader = new FileReader();
      reader.onload = (e) => {
        try {
          const jsonData = JSON.parse(e.target?.result as string);
          // Validate the JSON structure if needed
          if (validateJsonStructure(jsonData)) {
            templateData.value.template_data = jsonData;
          } else {
            console.error("Invalid JSON structure");
            // Handle the error (e.g., display an error message)
          }
        } catch (error) {
          console.error("Error parsing JSON:", error);
          // Handle the error (e.g., display an error message)
        }
      };
      reader.readAsText(file); // Read the file as text
    } else {
      // Handle the case where the file is undefined (e.g., display an error message)
      console.warn("No file selected");
    }
  } else {
    console.warn("Uploaded file is not a JSON file");
  }
};

const validateJsonStructure = (data: any[]) => {
  // Check if data is an array
  if (!Array.isArray(data)) {
    return false;
  }
  // Check if each item has the required properties
  return data.every((item) => {
    return (
      typeof item.query === "string" &&
      typeof item.response === "string" &&
      Array.isArray(item.filter) &&
      typeof item.source === "string" &&
      typeof item.block === "string" && // Assuming block is a string "yes" or "no"
      typeof item.category === "string"
    );
  });
};

const getTemplate = async (templateId: string) => {
  try {
    const response = await fetch(`/templates/${templateId}`);
    if (!response.ok) {
      throw new Error("Template not found");
    }
    const data = await response.json();
    templateData.value = data;
    console.log(data);
    if (props.templateId) {
      templateData.value.template_id = props.templateId;
    }
    if (!data.template_data) {
      templateData.value.template_data = [];
    }
    editMode.value = true;
  } catch (error) {
    // Handle error (e.g., display error message, redirect)
  }
};

const submitForm = async () => {
  loading.value = true;

  try {
    // Prepare data for API (assuming filter is a comma-separated string)
    const dataToSend = {
      ...templateData.value,
      template_data: templateData.value.template_data.map((item) => ({
        ...item,
        filter: typeof item.filter === "string" ? item.filter.split(",") : [], // Convert filter string to array
      })),
    };

    const response = await fetch(editMode.value ? "/update_template" : "/add_template", {
      method: editMode.value ? "PUT" : "POST",
      headers: {
        "Content-Type": "application/json",
        // Add authentication headers if needed
      },
      body: JSON.stringify(dataToSend),
    });

    if (!response.ok) {
      throw new Error("Failed to submit form");
    }

    const responseData = await response.json();
    console.log("Form submitted successfully:", responseData);

    // ... handle success (e.g., display success message, redirect)
  } catch (error) {
    console.error("Error submitting form:", error);
    // ... handle error
  } finally {
    loading.value = false;
    emit("close");
  }
};

const addItem = () => {
  templateData.value.template_data.push({
    query: "Enter your query",
    response: "",
    filter: "",
    source: "",
    block: false,
    category: "",
  });
};

const props = defineProps({
  templateId: {
    type: String,
    required: false,
  },
});

const removeItem = (index: number) => {
  templateData.value.template_data.splice(index, 1);
};

onMounted(() => {
  if (props.templateId) {
    getTemplate(props.templateId);
  } else {
    // Initialize for a new template
    templateData.value = {
      template_id: "",
      template_data: [],
      test_request: {
        body: { query: "{query}" },
        headers: {
          "Content-Type": "application/json",
        },
        method: "POST",
        url: "https://example.com/request",
      },
    };
  }
});

onUnmounted(() => {
  // ... Perform any cleanup if necessary
});
</script>
