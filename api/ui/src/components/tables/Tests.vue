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
  <n-table class="table-min-width" striped>
    <thead>
      <tr>
        <th></th>
        <th>Run ID</th>
        <th>Start Time</th>
        <th>End Time</th>
        <th>Status</th>
        <th>Progress</th>
        <th>Template ID</th>
      </tr>
    </thead>
    <tbody>
      <tr
        v-for="run in runs"
        :key="run.run_id"
        @click="openRunDetails(run.run_id)"
        class="pointer"
      >
        <td>
          <span v-if="run.status == 'Running'"><n-spin size="small" /></span>
          <span v-else-if="run.status == 'Not Started'"
            ><Icon :name="NotStartedIcon" :size="30"
          /></span>
          <span v-else><Icon :name="CompleteIcon" :size="30" /></span>
        </td>
        <td>{{ run.run_id }}</td>
        <td>{{ DateTimeFilter(run.start_time) }}</td>
        <td>{{ DateTimeFilter(run.end_time) }}</td>
        <td>{{ run.status }}</td>
        <td>{{ run.progress }}</td>
        <td>{{ run.template_id }}</td>
      </tr>
    </tbody>
  </n-table>
</template>

<script lang="ts" setup>
import { NTable, NSpin } from "naive-ui";
import { ref, onMounted, computed, watch } from "vue";
import { useRouter } from "vue-router";
import { format } from "date-fns"; // Import date-fns
import Icon from "@/components/common/Icon.vue";
const CompleteIcon = "carbon:checkmark-outline";
const NotStartedIcon = "carbon:pending";

// Define the Run interface
interface Run {
  run_id: string;
  status: string;
  progress: number;
  template_id: string;
  start_time: string;
  end_time: string;
  // ...other properties as needed
}

const router = useRouter();
const runs = ref<Run[]>([]); // Initialize runs as an empty array of Run type
const selectedRunId = ref<string | null>(null);

const openRunDetails = (runId: string) => {
  router.push({ name: "runDetails", params: { runId } });
};

const DateTimeFilter = (dateString: string) => {
  if (dateString) {
    const date = new Date(dateString); // Attempt to create a Date object
    if (isNaN(date.getTime())) {
      console.error("Invalid date string:", dateString);
      return "";
    }
    return format(date, "d/M/yy HH:mm");
  } else {
    return "";
  }
};

const hasRunningRuns = computed(() => runs.value.some((run) => run.status === "Running"));
const hasNotStartedRuns = computed(() =>
  runs.value.some((run) => run.status === "Not Started")
);

const fetchRuns = async () => {
  try {
    const response = await fetch("/runs");
    const data = await response.json();
    runs.value = data.runs as Run[];
  } catch (error) {
    console.error("Error fetching runs:", error);
  }
};

onMounted(() => {
  fetchRuns();
});

// Refresh every 10 seconds if there are running or not started runs
const refreshInterval = 4000; // 1 second in milliseconds
let refreshTimer: NodeJS.Timeout | undefined;
watch([hasRunningRuns, hasNotStartedRuns], (isRunning, isNotStarted) => {
  if (isRunning || isNotStarted) {
    if (!refreshTimer) {
      refreshTimer = setInterval(fetchRuns, refreshInterval);
    }
  } else if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = undefined;
  }
});
</script>

<style>
.pointer {
  cursor: pointer;
}
</style>
