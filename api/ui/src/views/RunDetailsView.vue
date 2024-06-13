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
    <h2>Run Details (Run ID: {{ runId }})</h2>
    <div class="filter-container">
      <n-input
        v-model:value="requestFilter"
        placeholder="Request Filter (e.g., body.query)"
        clearable
        @change="fetchRunDetails"
      />
      <n-input
        v-model:value="responseFilter"
        placeholder="Response Filter (e.g., response.output.text,assessment,status)"
        clearable
        @change="fetchRunDetails"
      />
      <n-input
        v-model:value="goldenResponsesFilter"
        placeholder="Golden Responses Filter (e.g., text)"
        clearable
        @change="fetchRunDetails"
      />
    </div>
    <n-spin :show="show">
      <n-table :data="testCases" class="table-min-width" striped>
        <thead>
          <tr>
            <th>Test Case ID</th>
            <th>Status</th>
            <th>Request</th>
            <th>Response</th>
            <th>Golden Answer</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="testCase in testCases" :key="testCase.id">
            <td>{{ testCase.id }}</td>
            <td>
              <span v-if="testCase.response.status === 'Failed'" style="color: red"
                >Failed</span
              >
              <span v-else>Passed</span>
            </td>
            <td>
              <pre class="wrap-text">{{ JSON.stringify(testCase.request, null, 2) }}</pre>
            </td>
            <td>
              <pre class="wrap-text" v-if="testCase.response.error">
              {{ testCase.response.error }}
            </pre
              >
              <pre class="wrap-text" v-else>
              {{ JSON.stringify(testCase.response, null, 2) }}
            </pre
              >
            </td>
            <td>
              <pre class="wrap-text">{{
                JSON.stringify(testCase.golden_response, null, 2)
              }}</pre>
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
import { NTable, NInput, NSpin } from "naive-ui";
import { useRoute } from "vue-router";
import { ref, onMounted, watch } from "vue";

interface TestCase {
  id: string;
  response: {
    status: string;
    error?: string;
  };
  request: any;
  golden_response: any;
}

const route = useRoute();
const runId = route.params.runId as string;
const testCases = ref<TestCase[]>([]);

const responseFilter = ref(
  "response.output.text,response.output.intent,assessment,status"
);
const requestFilter = ref("body.query");
const goldenResponsesFilter = ref("");
const show = ref(false);

const fetchRunDetails = () => {
  show.value = true;
  const filterString = `response_filter=${responseFilter.value}&request_filter=${requestFilter.value}&golden_responses_filter=${goldenResponsesFilter.value}`;
  fetch(`/run_status/${runId}?${filterString}`)
    .then((response) => response.json())
    .then((data) => {
      testCases.value = data.testCases as TestCase[];
      show.value = false;
    })
    .catch((error) => {
      console.error("Error fetching run details:", error);
      show.value = false;
    });
};

onMounted(() => {
  fetchRunDetails();
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
