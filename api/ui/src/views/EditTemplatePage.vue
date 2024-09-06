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
  <!-- Breadcrumb navigation for user orientation -->
  <Breadcrumbs :title="page.title"></Breadcrumbs>

  <div class="edit-template-page">
    <!-- EditTemplate component for template modification -->
    <!-- Pass the templateId as a string prop -->
    <edit-template :templateId="String(templateId)" @close="handleClose" @updateSuccess="handleUpdateSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { useRoute, useRouter } from 'vue-router';
import EditTemplate from '@/components/EditTemplate.vue';
import { ref } from 'vue';
import Breadcrumbs from '@/components/shared/Breadcrumbs.vue';

// Page title reactive reference
const page = ref({ title: 'Edit Template' });

// Get route information and router instance
const route = useRoute();
const router = useRouter();

// Extract templateId from route parameters
const templateId = route.params.templateId;

/**
 * Handles the close event from the EditTemplate component.
 * Navigates back to the previous page in the history.
 */
const handleClose = () => {
  router.back();
};

/**
 * Handles the successful update event from the EditTemplate component.
 * Redirects the user to the templates list page.
 */
const handleUpdateSuccess = () => {
  router.push({ name: 'templates' });
};
</script>
