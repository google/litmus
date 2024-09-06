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
  <!-- Breadcrumb navigation for better orientation within the application -->
  <Breadcrumbs :title="page.title"></Breadcrumbs>

  <!-- Main content area of the page -->
  <div class="page">
    <!-- Container for the list of proxies, allowing horizontal scrolling -->
    <div class="components-list">
      <n-scrollbar x-scrollable style="width: 100%">
        <!-- Proxies component to display the list of proxies -->
        <Proxies class="table-min-width" striped />
      </n-scrollbar>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { NScrollbar } from 'naive-ui';
import Proxies from '@/components/tables/Proxies.vue';
import { ref } from 'vue';
import Breadcrumbs from '@/components/shared/Breadcrumbs.vue';

// Data for the page title, used by the Breadcrumbs component
const page = ref({ title: 'Proxies' });
</script>

<style scoped lang="scss">
.components-list {
  // Ensure the list takes up the full width and adjusts for columns
  grid-template-columns: none;
}

.table-min-width {
  // Set a minimum width for the Proxies table to ensure readability
  min-width: 480px;
}
</style>
