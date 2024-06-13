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
  <Breadcrumbs :title="page.title" :breadcrumbs="breadcrumbs"></Breadcrumbs>
  <div class="page">
    <div class="components-list">
      <n-scrollbar x-scrollable style="width: 100%">
        <run />
      </n-scrollbar>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { NScrollbar } from "naive-ui";
import Run from "@/components/SubmitRun.vue";
import { ref } from "vue";
import Breadcrumbs from "@/components/shared/Breadcrumbs.vue";

const page = ref({ title: "Start New Run" });
const breadcrumbs = ref([
  {
    title: "Start New Run",
    disabled: true,
    href: "#",
  },
]);
</script>

<style scoped lang="scss">
.components-list {
  grid-template-columns: none;
}

.table-min-width {
  min-width: 480px;
}
</style>
