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
    Displays the list of tests.

    @slot default The default slot for the component
   -->
  <div>
    <!-- Breadcrumbs component with dynamic title -->
    <Breadcrumbs :title="page.title"></Breadcrumbs>

    <div class="page">
      <div class="components-list">
        <!-- 
          Provides a scrollable container for the Tests table.
          The Tests table is set to take up 100% of the available width. 
         -->
        <n-scrollbar x-scrollable style="width: 100%">
          <!-- 
            Tests table component.
            The table is set to have a minimum width to prevent content from wrapping. 
          -->
          <Tests class="table-min-width" striped />
        </n-scrollbar>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
// Import necessary components and libraries
import { NScrollbar } from 'naive-ui';
import Tests from '@/components/tables/Tests.vue';
import { ref } from 'vue';
import Breadcrumbs from '@/components/shared/Breadcrumbs.vue';

// Define page title using ref for reactivity
const page = ref({ title: 'Tests' });
</script>

<style scoped lang="scss">
// Component-specific styles
.components-list {
  grid-template-columns: none;
}

// Style for minimum width of the table
.table-min-width {
  min-width: 480px;
}
</style>
