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
  <!-- UI Test Runs Table component for displaying and managing test runs. -->
  <n-spin :show="show">
    <!-- Display a loading spinner while fetching data. -->
    <n-table class="table-min-width" striped>
      <!-- Define the table header. -->
      <thead>
        <tr>
          <th></th>
          <th>Run ID</th>
          <th>Start Time</th>
          <th>End Time</th>
          <th>Status</th>
          <th>Progress</th>
          <th>Template ID</th>
          <th>Restart</th>
          <th>Delete</th>
        </tr>
      </thead>
      <!-- Populate the table body with test run data. -->
      <tbody>
        <tr v-for="run in runs" :key="run.run_id" class="pointer">
          <!-- Display a status icon based on the run status. -->
          <td>
            <span v-if="run.status == 'Running'"><VueSpinner size="30" /></span>
            <span v-else-if="run.status == 'Not Started'"><Icon :name="NotStartedIcon" :size="30" /></span>
            <span v-else style="color: green"><Icon :name="CompleteIcon" :size="30" /></span>
          </td>
          <!-- Display other run details in clickable table cells. -->
          <td @click="openRunDetails(run.run_id)">{{ run.run_id }}</td>
          <td @click="openRunDetails(run.run_id)">{{ DateTimeFilter(run.start_time) }}</td>
          <td @click="openRunDetails(run.run_id)">{{ DateTimeFilter(run.end_time) }}</td>
          <td @click="openRunDetails(run.run_id)">{{ run.status }}</td>
          <td @click="openRunDetails(run.run_id)">{{ run.progress }}</td>
          <td @click="openRunDetails(run.run_id)">{{ run.template_id }}</td>
          <!-- Provide buttons for restarting and deleting runs. -->
          <td>
            <Icon @click="triggerRestart(run.template_id, run.run_id)" :name="RestartIcon" :size="25" />
          </td>
          <td>
            <Icon @click="triggerDelete(run.run_id)" :name="RemoveIcon" :size="25" />
          </td>
        </tr>
      </tbody>
    </n-table>
  </n-spin>
</template>

<script lang="ts" setup>
import { NTable, NSpin, useMessage } from 'naive-ui';
import { ref, onMounted, computed, watch } from 'vue';
import { useRouter } from 'vue-router';
import { format } from 'date-fns'; // Import date-fns for date formatting.
import Icon from '@/components/common/Icon.vue';
import { VueSpinner } from 'vue3-spinners';

// Define icon names for different run statuses.
const CompleteIcon = 'carbon:checkmark-outline';
const RestartIcon = 'carbon:restart';
const NotStartedIcon = 'carbon:pending';
const RemoveIcon = 'carbon:trash-can';

// Define the Run interface for type safety.
interface Run {
  run_id: string;
  status: string;
  progress: number;
  template_id: string;
  start_time: string;
  end_time: string;
}

// Initialize Vue components and variables.
const router = useRouter();
const message = useMessage();
const runs = ref<Run[]>([]); // Store fetched run data.
const show = ref(false); // Control the loading spinner visibility.

// Function to navigate to the run details page.
const openRunDetails = (runId: string) => {
  router.push({ name: 'runDetails', params: { runId } });
};

// Function to format date strings to a user-friendly format.
const DateTimeFilter = (dateString: string) => {
  if (dateString) {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      console.error('Invalid date string:', dateString);
      return '';
    }
    return format(date, 'd/M/yy HH:mm');
  } else {
    return '';
  }
};

// Computed properties to check for running or not started runs.
const hasRunningRuns = computed(() => runs.value.some((run) => run.status === 'Running'));
const hasNotStartedRuns = computed(() => runs.value.some((run) => run.status === 'Not Started'));

// Function to fetch run data from the backend API.
const fetchRuns = async () => {
  show.value = true; // Show the loading spinner.
  try {
    const response = await fetch('/runs');
    const data = await response.json();
    runs.value = data.runs as Run[];
    show.value = false; // Hide the spinner after fetching data.
  } catch (error) {
    console.error('Error fetching runs:', error);
    show.value = false;
  }
};

// Function to fetch run data in the background.
const fetchRunsBackground = async () => {
  try {
    const response = await fetch('/runs');
    const data = await response.json();
    runs.value = data.runs as Run[];
  } catch (error) {
    console.error('Error fetching runs:', error);
  }
};

// Function to trigger a run restart on the backend.
const triggerRestart = async (templateId: string, runId: string) => {
  show.value = true;
  try {
    const response = await fetch('/invoke_run', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        run_id: runId,
        template_id: templateId
      })
    });
    if (response.ok) {
      message.success('Test run submitted successfully!');
      fetchRuns();
      show.value = false;
    } else {
      const data = await response.json();
      message.error(data.error);
      show.value = false;
    }
  } catch (error) {
    message.error('Error submitting run.');
    show.value = false;
  }
};

// Function to trigger a run deletion on the backend.
const triggerDelete = async (runId: string) => {
  show.value = true;
  try {
    const response = await fetch('/delete_run/' + runId, {
      method: 'DELETE'
    });
    if (response.ok) {
      message.success('Run ' + runId + ' deleted successfully!');
      fetchRuns();
      show.value = false;
    } else {
      const data = await response.json();
      message.error(data.error);
      show.value = false;
    }
  } catch (error) {
    message.error('Error deleting run.');
    show.value = false;
  }
};

// Fetch initial run data when the component is mounted.
onMounted(() => {
  fetchRuns();
});

// Set up a timer to refresh run data periodically.
const refreshInterval = 3000; // 3 seconds in milliseconds
let refreshTimer: NodeJS.Timeout | undefined;
watch([hasRunningRuns, hasNotStartedRuns], (isRunning, isNotStarted) => {
  if (isRunning || isNotStarted) {
    if (!refreshTimer) {
      refreshTimer = setInterval(fetchRunsBackground, refreshInterval);
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
