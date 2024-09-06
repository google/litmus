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

<script lang="ts" setup>
import { onMounted } from 'vue';
import { Chart, registerables } from 'chart.js';

// Register Chart.js components
Chart.register(...registerables);

// Interface for individual data points
interface DataPoint {
  // Using index signature to allow for flexible data structure
  [key: string]: any;
}

// Define props expected by this component
const props = defineProps<{
  // Array of data points to be displayed on the chart
  data: DataPoint[];
  // Unique ID for the chart canvas element
  chartId: string;
  // Key within each data point object to extract the value for the Y-axis
  responseKey: string;
}>();

// Configuration options for the Chart.js Line chart
const chartOptions = {
  // Customize your chart options here
  // Example: Ensuring the Y-axis starts at zero
  scales: {
    y: {
      beginAtZero: true
    }
  }
};

// Function executed when the component is mounted in the DOM
onMounted(() => {
  // Retrieve the canvas element where the chart will be rendered
  const ctx = document.getElementById(props.chartId) as HTMLCanvasElement;

  // Transform the raw data into a format suitable for Chart.js
  const dataPoints = props.data.map((items: DataPoint) => ({
    // Extract the 'start_time' property from each data point and convert it to a Date object for the X-axis
    x: new Date(items.start_time),
    // Extract the value associated with the 'responseKey' from each data point for the Y-axis
    y: items.data[props.responseKey] // Optional chaining handles cases where 'data' or 'responseKey' might be undefined
  }));

  // Extract and format X-axis labels from dataPoints
  const labels = dataPoints.map((items) => items.x.toLocaleString());

  // Create a new Chart.js Line chart instance
  new Chart(ctx, {
    type: 'line',
    data: {
      labels, // X-axis labels
      datasets: [
        {
          // Label for the dataset, dynamically set to the value of 'responseKey'
          label: props.responseKey,
          data: dataPoints // Data points for the chart
          // Customize line style and color here
        }
      ]
    },
    options: chartOptions // Apply the defined chart options
  });
});
</script>
