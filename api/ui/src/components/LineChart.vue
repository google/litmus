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
import { onMounted, ref } from "vue";
import { Chart, registerables } from "chart.js";

Chart.register(...registerables);


interface DataPoint {
  [key: string]: any;
}

const props = defineProps<{
  data: DataPoint[];
  chartId: string;
  responseKey: string;
}>();

const chartOptions = {
  // Customize your chart options here
  // Example:
  scales: {
    y: {
      beginAtZero: true,
    },
  },
};

onMounted(() => {
  const ctx = document.getElementById(props.chartId) as HTMLCanvasElement;
    const dataPoints = props.data.map((items: DataPoint) => ({
        x: new Date(items.start_time),
        y: items.data[props.responseKey], // Optional chaining
    }));
  const labels = dataPoints.map(items => items.x.toLocaleString());

  new Chart(ctx, {
    type: "line",
    data: {
      labels,
      datasets: [
        {
          label: props.responseKey, // Dynamic label based on responseKey
          data: dataPoints,
          // Customize line style and color here
        },
      ],
    },
    options: chartOptions,
  });
});
</script>