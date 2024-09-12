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
    <!-- Loading spinner for the template table -->
    <n-spin :show="show">
      <!-- Table to display a list of template IDs -->
      <n-table class="table-min-width" striped>
        <thead>
          <tr>
            <th>Template ID</th>
            <th>Actions</th>
          </tr>
        </thead>
        <tbody>
          <!-- Iterate over the templateIds array to display each template ID -->
          <tr v-for="template in templates" :key="template.template_id">
            <!-- Display the templateId -->
            <td>{{ template.template_id }}</td>
            <td>
              <!-- Button to navigate to the comparison page for the specific templateId -->
              <v-btn variant="flat" color="secondary" class="mt-2" @click="navigateToComparePage(template.template_id)"
                >Compare Tests</v-btn
              >
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-spin>
    <n-divider />
  </div>
</template>

<script lang="ts" setup>
import { NTable, NDivider, NSpin } from 'naive-ui';
import { ref, onMounted } from 'vue';
import { useRouter } from 'vue-router';

// Define a ref to store an array of template IDs
const templates = ref<{ template_id: string; template_type: string | null }[]>([]);

// Get the router instance
const router = useRouter();

// Reactive variable controlling the visibility of the loading spinner for the table
const show = ref(false);

/**
 * Navigates to the comparison page for the given template ID.
 *
 * @param {string} templateId - The ID of the template to compare.
 */
const navigateToComparePage = (templateId: string) => {
  // Push the 'compare' route with the templateId as a parameter
  router.push({ name: 'compare', params: { templateId } });
};

/**
 * Fetches the list of UI Templates from the backend.
 */
const fetchTemplates = () => {
  // Show the loading spinner
  show.value = true;
  // Fetch the templates from the backend API
  fetch('/templates?type=Test%20Run')
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

// Fetch the templates when the component is mounted
onMounted(() => {
  fetchTemplates();
});
</script>
