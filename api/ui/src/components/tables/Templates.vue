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
    <n-table class="table-min-width" striped>
      <thead>
        <tr>
          <th>Template ID</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="templateId in templateIds" :key="templateId">
          <td>{{ templateId }}</td>
          <td>
            <v-btn
              variant="flat"
              color="secondary"
              class="mt-2"
              @click="navigateToComparePage(templateId)"
              >View Statistics</v-btn
            >
            &nbsp;
            <v-btn
              variant="flat"
              color="primary"
              class="mt-2"
              @click="navigateToEditPage(templateId)"
              >Edit</v-btn
            >
            &nbsp;
            <v-btn
              variant="flat"
              color="accent"
              class="mt-2"
              @click="deleteTemplate(templateId)"
              >Delete</v-btn
            >
          </td>
        </tr>
      </tbody>
    </n-table>
  </div>
  <n-divider />
  <n-space>
    <v-btn variant="flat" color="primary" class="mt-2" @click="openAddTemplateModal"
      >Add Template</v-btn
    >
  </n-space>
</template>

<script lang="ts" setup>
import {
  NTable,
  NButton,
  NModal,
  useMessage,
  NPopconfirm,
  NDivider,
  NSpace,
} from "naive-ui";
import { ref, onMounted } from "vue";
import { useRouter } from "vue-router";

const templateIds = ref<string[]>([]);
const message = useMessage();

const router = useRouter();

const navigateToEditPage = (templateId: string) => {
  router.push({ name: "editTemplate", params: { templateId } });
};

const navigateToComparePage = (templateId: string) => {
  router.push({ name: "compare", params: { templateId } });
};

const openAddTemplateModal = () => {
  router.push({ name: "addTemplate" });
};

const deleteTemplate = async (templateId: string) => {
  try {
    const response = await fetch(`/delete_template/${templateId}`, {
      method: "DELETE",
      // ... add authentication headers if needed
    });
    if (response.ok) {
      // Template deleted successfully
      // Update the UI to remove the template from the list
      // ...
      fetchTemplates();
      message.success("Template deleted successfully"); // Assuming you have a message component
    } else {
      // Handle error
      const errorData = await response.json();
      message.error(`Error deleting template: ${errorData.error}`);
    }
  } catch (error) {
    // Handle unexpected errors
    console.error("Error deleting template:", error);
    message.error("An unexpected error occurred while deleting the template.");
  }
};

const fetchTemplates = () => {
  fetch("/templates")
    .then((response) => response.json())
    .then((data) => {
      templateIds.value = data.template_ids;
    })
    .catch((error) => {
      console.error("Error fetching templates:", error);
    });
};

onMounted(() => {
  fetchTemplates();
});
</script>
