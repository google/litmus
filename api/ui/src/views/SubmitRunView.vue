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
  <!-- Breadcrumbs for navigation -->
  <Breadcrumbs :title="page.title"></Breadcrumbs>

  <!-- Page container -->
  <div class="page">
    <!-- Container for the components list -->
    <div class="components-list">
      <!-- Scrollbar for the Run component -->
      <n-scrollbar x-scrollable style="width: 100%">
        <!-- Run component for submitting and viewing test runs -->
        <run />
      </n-scrollbar>
    </div>
  </div>
</template>

<script lang="ts" setup>
// Import necessary components and libraries
import { NScrollbar } from 'naive-ui';
import Run from '@/components/SubmitRun.vue';
import { ref } from 'vue';
import Breadcrumbs from '@/components/shared/Breadcrumbs.vue';

// Define a reactive object for the page title
const page = ref({ title: 'Start New Run' });
</script>

<style scoped lang="scss">
// Styles for the component list
.components-list {
  // Use grid layout with no predefined columns
  grid-template-columns: none;
}

// Styles for the table minimum width (unused in this template)
.table-min-width {
  min-width: 480px;
}
</style>
