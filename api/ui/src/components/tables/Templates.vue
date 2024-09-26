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
    <!-- Spin Loading indicator while data is being fetched -->
    <n-spin :show="show">
      <!-- Table to display UI Templates -->
      <n-table class="table-min-width" striped>
        <thead>
          <tr>
            <th>Template ID</th>
            <th>Type</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <!-- Iterate over templates to display each template -->
          <tr v-for="template in templates" :key="template.template_id">
            <!-- Display template ID -->
            <td>{{ template.template_id }}</td>
            <!-- Display template type -->
            <td>{{ template.template_type || 'N/A' }}</td>
            <td>
              <!-- Edit button to navigate to Edit Template page -->
              <v-btn variant="flat" color="primary" class="mt-2" @click="navigateToEditPage(template.template_id)">Edit</v-btn>
               
              <!-- Delete button to delete a template -->
              <v-btn variant="flat" color="accent" class="mt-2" @click="deleteTemplate(template.template_id)">Delete</v-btn>
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-spin>
    <!-- Divider for visual separation -->
    <n-divider />
    <!-- Space component for spacing -->
    <n-space>
      <!-- Button to open Add Template Modal -->
      <v-btn variant="flat" color="primary" class="mt-2" @click="openAddTemplateModal">Add Template</v-btn>
    </n-space>
  </div>
</template>

<script lang="ts" setup>
import { NTable, useMessage, NDivider, NSpace, NSpin } from 'naive-ui';
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';

// Define a ref to store templates, initialized as an empty array
const templates = ref<{ template_id: string; template_type: string | null }[]>([]);
// Access the message instance from Naive UI
const message = useMessage();

// Access the router instance for navigation
const router = useRouter();
// Define a ref to control the loading spinner visibility
const show = ref(false);

/**
 * Navigates to the Edit Template page with the given templateId.
 * @param {string} templateId - The ID of the template to edit.
 */
const navigateToEditPage = (templateId: string) => {
  router.push({ name: 'editTemplate', params: { templateId } });
};

/**
 * Navigates to the Add Template page.
 */
const openAddTemplateModal = () => {
  router.push({ name: 'addTemplate' });
};

/**
 * Deletes a UI Template with the given templateId.
 * @param {string} templateId - The ID of the template to delete.
 */
const deleteTemplate = async (templateId: string) => {
  // Show the loading spinner
  show.value = true;
  try {
    // Send a DELETE request to the backend to delete the template
    const response = await fetch(`/templates/${templateId}`, {
      method: 'DELETE'
    });

    if (response.ok) {
      // If the deletion was successful, fetch the updated list of templates
      fetchTemplates();
      // Show a success message
      message.success('Template deleted successfully');
    } else {
      // If there was an error, parse the error data from the response
      const errorData = await response.json();
      // Show an error message with the error details
      message.error(`Error deleting template: ${errorData.error}`);
      // Hide the loading spinner
      show.value = false;
    }
  } catch (error) {
    // Log any unexpected errors during the deletion process
    console.error('Error deleting template:', error);
    // Show a generic error message
    message.error('An unexpected error occurred while deleting the template.');
    // Hide the loading spinner
    show.value = false;
  }
};

/**
 * Fetches the list of UI Templates from the backend.
 */
const fetchTemplates = () => {
  // Show the loading spinner
  show.value = true;
  // Fetch the templates from the backend API
  fetch('/templates')
    .then((response) => response.json())
    .then((data) => {
      // Update the templates ref with the fetched data
      templates.value = data.templates;
      // Hide the loading spinner after data is fetched
      show.value = false;
    })
    .catch((error) => {
      // Log any errors encountered while fetching templates
      console.error('Error fetching templates:', error);
      // Hide the loading spinner in case of an error
      show.value = false;
    });
};

// Fetch templates when the component is mounted
onMounted(() => {
  fetchTemplates();
});
</script>
