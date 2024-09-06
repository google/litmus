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
  <!-- Breadcrumb navigation for better user experience -->
  <Breadcrumbs :title="page.title"></Breadcrumbs>

  <!-- Container for the Add Template page -->
  <div class="edit-template-page">
    <!-- EditTemplate component for adding a new template -->
    <!-- Events are emitted upon closing the component or successful update -->
    <edit-template @close="handleClose" @updateSuccess="handleUpdateSuccess" />
  </div>
</template>

<script lang="ts" setup>
import { useRouter } from 'vue-router';
import EditTemplate from '@/components/EditTemplate.vue';
import { ref } from 'vue';
import Breadcrumbs from '@/components/shared/Breadcrumbs.vue';

// Page title reactive reference
const page = ref({ title: 'Add Template' });

// Router instance for navigation
const router = useRouter();

/**
 * Handles the close event emitted by EditTemplate component.
 * Navigates back to the previous page in the history.
 */
const handleClose = () => {
  router.back();
};

/**
 * Handles the updateSuccess event emitted by EditTemplate component.
 * Navigates to the 'templates' route upon successful template creation.
 */
const handleUpdateSuccess = () => {
  router.push({ name: 'templates' });
};
</script>
