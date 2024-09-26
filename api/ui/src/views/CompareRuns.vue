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
    <!-- Breadcrumb for navigation -->
    <v-row class="page-breadcrumb mb-0 mt-n2">
      <v-col cols="12" md="12">
        <v-card elevation="0" variant="text">
          <v-row no-gutters class="align-center">
            <v-col sm="12">
              <h3 class="text-h3 mt-1 mb-0">Template Results (Template ID: {{ templateId }})</h3>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>

    <!-- Filter inputs for requests and responses -->
    <div class="filter-container">
      <n-input v-model:value="requestFilter" placeholder="Request Filter (e.g., body.query)" clearable @change="fetchData" />
      <n-input
        v-model:value="responseFilter"
        placeholder="Response Filter (e.g., assessment.similarity,response.output.intent)"
        clearable
        @change="fetchData"
      />
    </div>

    <!-- Toggle switch to show/hide charts -->
    <div class="toggle-chart">
      Toggle Graph
      <n-switch v-model:value="showChart" size="large"> Show Charts </n-switch>
    </div>

    <n-divider></n-divider>

    <!-- Loading spinner while data is being fetched -->
    <n-spin :show="show">
      <!-- Table to display filtered run data -->
      <n-table :data="Object.entries(filteredData)" class="table-min-width" striped>
        <thead>
          <tr>
            <th>Request</th>
            <th>Data</th>
          </tr>
        </thead>
        <tbody>
          <!-- Display message if no results are found -->
          <tr v-if="Object.entries(filteredData).length == 0">
            <td colspan="2">No Results</td>
          </tr>
          <!-- Iterate over filtered data and display each run -->
          <tr v-for="([key, value], index) in Object.entries(filteredData)" :key="key">
            <!-- Display request key -->
            <td>{{ key }}</td>
            <!-- Conditionally render chart or preformatted JSON data -->
            <td v-if="showChart">
              <canvas :id="`x-chart-${index}`"></canvas>
              <LineChart :key="key" :chart-id="`x-chart-${index}`" :data="value as ChartData[]" :response-key="responseFilter" />
            </td>
            <td v-else>
              <pre class="wrap-text">{{ JSON.stringify(value, null, 2) }}</pre>
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
import { NTable, NInput, NSwitch, NDivider, NSpin } from 'naive-ui';
import { useRoute } from 'vue-router';
import { ref, onMounted, watch } from 'vue';
import LineChart from '@/components/LineChart.vue';

// Interface for run data structure
interface RunData {
  data: {
    [key: string]: any;
  };
  end_time: string;
  run_id: string;
  start_time: string;
}

// Interface for response data structure
interface ResponseData {
  [requestText: string]: RunData[];
}

// Interface for chart data structure
interface ChartData {
  start_time: string;
  [key: string]: any;
}

// Interface for filtered data structure
interface FilteredData {
  [key: string]: ChartData[];
}

// Get route parameters
const route = useRoute();
const templateId = route.params.templateId as string;

// Reactive variables
const responseData = ref<ResponseData>({}); // Store fetched response data
const responseFilter = ref('assessment.similarity'); // Store response filter value
const requestFilter = ref('body.query'); // Store request filter value
const showChart = ref(false); // Control chart visibility
const show = ref(false); // Control loading spinner visibility

/**
 * Fetches run data from the API based on filter values.
 */
const fetchData = () => {
  show.value = true; // Show loading spinner
  const filterString = `response_filter=${responseFilter.value}&request_filter=${requestFilter.value}`; // Construct filter query string
  fetch(`/runs/all_results/${templateId}?${filterString}`)
    .then((response) => response.json())
    .then((data) => {
      responseData.value = data;
      show.value = false; // Hide loading spinner
    })
    .catch((error) => {
      console.error('Error fetching run results:', error);
      show.value = false; // Hide loading spinner on error
    });
};

/**
 * Fetches default filter values from the API.
 */
const fetchRunFields = () => {
  fetch(`/templates/${templateId}`)
    .then((response) => response.json())
    .then((data) => {
      requestFilter.value = data.template_input_field; // Set default request filter
      fetchData(); // Fetch data with default filters
    })
    .catch((error) => {
      console.error('Error fetching run details:', error);
    });
};

// Reactive variable for filtered data
const filteredData = ref<FilteredData>({});

// Watcher to update filteredData whenever responseData or responseFilter changes
watch([responseData, responseFilter], () => {
  // Filter responseData based on responseFilter value
  filteredData.value = Object.fromEntries(
    Object.entries(responseData.value).map(([key, value]) => {
      const filteredRuns = value.filter((run) => {
        const data = run.data;
        const filterParts = responseFilter.value.split(',');
        return filterParts.every((part) => data.hasOwnProperty(part));
      });
      return [key, filteredRuns];
    })
  );
});

// Fetch initial data on component mount
onMounted(() => {
  fetchRunFields();
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

.toggle-chart {
  display: flex;
  align-items: center;
  gap: 10px;
}
</style>
