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
  <!-- 
    Displays a list of Litmus Templates.

    Templates are reusable test definitions used to verify the behavior 
    of HTTP requests and responses, including those from LLMs. 
  -->

  <!-- Breadcrumbs component for page navigation -->
  <Breadcrumbs :title="page.title"></Breadcrumbs>

  <!-- Page container -->
  <div class="page">
    <!-- 
      Components list container with horizontal scrollbar.

      The 'components-list' class applies specific styles to 
      the container, including grid display properties 
      and horizontal scrollbar. 
    -->
    <div class="components-list">
      <!-- 
        Naive UI Scrollbar component to enable horizontal 
        scrolling for the templates table.

        The 'x-scrollable' directive enables horizontal scrolling.
      -->
      <n-scrollbar x-scrollable style="width: 100%">
        <!-- 
          Templates table component with a minimum width class.

          The 'Templates' component displays the table of Litmus templates.
          The 'table-min-width' class sets a minimum width for the table
          to ensure readability and prevent content from being truncated. 
        -->
        <Templates class="table-min-width" striped />
      </n-scrollbar>
    </div>
  </div>
</template>

<script lang="ts" setup>
// Import necessary components and utilities
import { NScrollbar } from 'naive-ui';
import Templates from '@/components/tables/Templates.vue';
import { ref } from 'vue';
import Breadcrumbs from '@/components/shared/Breadcrumbs.vue';

// Page title data
const page = ref({ title: 'Templates' });
</script>

<style scoped lang="scss">
// Styles for components list container
.components-list {
  // Set grid template columns to none to use the default grid layout
  grid-template-columns: none;
}

// Styles for table minimum width
.table-min-width {
  // Set minimum width to 480px to prevent content truncation
  min-width: 480px;
}
</style>
