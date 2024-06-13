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
    <h2>Template Results (Template ID: {{ templateId }})</h2>
    <div class="filter-container">
      <n-input
        v-model:value="requestFilter"
        placeholder="Request Filter (e.g., body.query)"
        clearable
        @change="fetchData"
      />
      <n-input
        v-model:value="responseFilter"
        placeholder="Response Filter (e.g., assessment.similarity,response.output.intent)"
        clearable
        @change="fetchData"
      />
    </div>
    Toggle Graph
    <n-switch v-model:value="showChart" size="large"> Show Charts </n-switch>
    <n-divider></n-divider>
    <n-spin :show="show">
      <n-table :data="Object.entries(filteredData)" class="table-min-width" striped>
        <thead>
          <tr>
            <th>Request</th>
            <th>Data</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="([key, value], index) in Object.entries(filteredData)" :key="key">
            <td>{{ key }}</td>
            <td v-if="showChart">
              <canvas :id="`x-chart-${index}`">{{ key }}</canvas>
              <LineChart
                :key="key"
                :chart-id="`x-chart-${index}`"
                :data="value as ChartData[]"
                :response-key="responseFilter"
              />
            </td>
            <td>
              <pre class="wrap-text" v-if="!showChart">
              {{ JSON.stringify(value, null, 2) }}
            </pre
              >
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
import { NTable, NInput, NSwitch, NDivider, NSpin } from "naive-ui";
import { useRoute } from "vue-router";
import { ref, onMounted, watch, computed } from "vue";
import LineChart from "@/components/LineChart.vue";

interface RunData {
  data: {
    [key: string]: any; // Adjust based on your actual data structure
  };
  end_time: string;
  run_id: string;
  start_time: string;
}

interface ResponseData {
  [requestText: string]: RunData[];
}

interface ChartData {
  start_time: string;
  [key: string]: any; // To accommodate the dynamic responseKey
}

interface FilteredData {
  [key: string]: ChartData[];
}

const route = useRoute();
const templateId = route.params.templateId as string;
const responseData = ref<ResponseData>({});
const responseFilter = ref("assessment.similarity");
const requestFilter = ref("body.query");
const showChart = ref(false);
const show = ref(false);

const fetchData = () => {
  show.value = true;
  const filterString = `response_filter=${responseFilter.value}&request_filter=${requestFilter.value}`;
  fetch(`/all_run_results/${templateId}?${filterString}`)
    .then((response) => response.json())
    .then((data) => {
      responseData.value = data;
      show.value = false;
    })
    .catch((error) => {
      console.error("Error fetching run results:", error);
      show.value = false;
    });
};

const filteredData = ref<FilteredData>({});

watch([responseData, responseFilter], () => {
  filteredData.value = Object.fromEntries(
    Object.entries(responseData.value).map(([key, value]) => {
      const filteredRuns = value.filter((run) => {
        const data = run.data;
        const filterParts = responseFilter.value.split(",");
        return filterParts.every((part) => data.hasOwnProperty(part));
      });
      return [key, filteredRuns];
    })
  );
});

onMounted(() => {
  fetchData();
});
</script>

<style>
.wrap-text {
  white-space: pre-wrap;
  word-break: break-all;
  min-width: 20em;
}

.filter-container {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
}
</style>
