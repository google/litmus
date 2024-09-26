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
    <!-- Page Title and Breadcrumb: Dynamically displays title based on selected data source -->
    <v-row class="page-breadcrumb mb-0 mt-n2">
      <v-col cols="12" md="12">
        <v-card elevation="0" variant="text">
          <v-row no-gutters class="align-center">
            <v-col sm="12">
              <h3 class="text-h3 mt-1 mb-0">{{ selectedDataSource === 'proxy/data' ? 'Proxy Data' : 'Litmus Data' }}</h3>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>

    <!-- Filter and Action Buttons Container -->
    <div class="filter-container">
      <!-- Data Source Selection: Switch to choose between Litmus and Proxy data -->
      <n-select
        class="litmus-data-selector"
        v-model:value="selectedDataSource"
        :options="[
          { label: 'Litmus Data', value: 'proxy/litmus_data	' },
          { label: 'Proxy Data', value: 'proxy/data' }
        ]"
      />
      <!-- Date picker for filtering by date (uses Unix timestamp internally) -->
      <n-date-picker v-model:value="selectedDate" type="date" />
      <!-- Input field for filtering by context -->
      <n-input v-model:value="contextFilter" placeholder="Context Filter" clearable />
      <!-- Button to fetch data based on filters -->
      <v-btn variant="flat" color="primary" @click="fetchData"> Fetch Data </v-btn>
      <!-- Button to export data to a CSV file (disabled if no data is available) -->
      <v-btn variant="flat" color="primary" @click="exportCSV" :disabled="!showData"> Export Data to CSV </v-btn>
      <!-- Toggle button to show/hide the API response section -->
      <n-switch v-model:value="showApiResponse" />
      <span>Show API Response</span>
    </div>

    <!-- Data Table: Displays data, loading spinner, or error messages -->
    <n-spin :show="showLoading">
      <!-- Error message section (shown if there's an error fetching data) -->
      <div v-if="showError" class="error-message">
        <n-icon color="red" size="20">
          <WarningOutline />
        </n-icon>
        <span v-html="errorMessage"></span>
      </div>
      <!-- Data table section (shown if data is available) -->
      <div v-else-if="showData">
        <!-- Field Selection Collapse Panel: Allows users to select fields to display -->
        <n-collapse v-model:expanded="collapseExpanded">
          <n-collapse-item title="Select Fields" name="select-fields">
            <div class="checkbox-container">
              <!-- Dynamic checkboxes for each field, allowing users to toggle visibility -->
              <n-checkbox v-for="field in availableFields" :key="field" v-model:checked="selectedFields[field]">
                <span
                  :style="{
                    fontWeight: isImportantField(field) ? 'bold' : 'normal'
                  }"
                >
                  {{ field }}
                </span>
              </n-checkbox>
            </div>
          </n-collapse-item>
        </n-collapse>

        <!-- Data Table for Selected Fields -->
        <div class="table-container">
          <n-table :data="sortedData" :single-line="false" size="small" striped>
            <thead>
              <tr>
                <!-- Table headers, sortable by clicking -->
                <th v-for="field in visibleFields" :key="field" @click="sortTable(field)" class="sortable-header">
                  {{ field }}
                  <!-- Sort direction indicator -->
                  <span v-if="sortField === field">
                    <n-icon :component="sortDirection === 'asc' ? CaretUpOutline : CaretDownOutline" />
                  </span>
                </th>
              </tr>
            </thead>
            <tbody>
              <!-- Render table rows with sorted and filtered data -->
              <tr v-for="(record, index) in sortedData" :key="index">
                <td v-for="field in visibleFields" :key="field">
                  <!-- Display object values as JSON -->
                  <template v-if="typeof record[field] === 'object'">
                    {{ JSON.stringify(record[field], null, 2) }}
                  </template>
                  <!-- Display primitive values directly -->
                  <template v-else>
                    {{ record[field] }}
                  </template>
                </td>
              </tr>
            </tbody>
          </n-table>
        </div>
        <!-- Display API Response (Optional): Shows the raw API response for debugging -->
        <div v-if="showApiResponse" class="api-response">
          <h3>API Response:</h3>
          <pre>{{ apiResponse }}</pre>
        </div>
      </div>
      <!-- "No data available" message (shown if no data is fetched and there's no error) -->
      <div v-else-if="!showLoading">No data available.</div>
    </n-spin>
  </div>
</template>

<script lang="ts" setup>
// Import necessary components and functions from UI libraries
import { NTable, NInput, NSpin, NDatePicker, NCheckbox, NIcon, NCollapse, NCollapseItem, NSwitch, NSelect } from 'naive-ui';
import { ref, onMounted, computed, watch } from 'vue';
// Import icons from icon libraries
import { CaretUpOutline, CaretDownOutline, WarningOutline } from '@vicons/ionicons5';
// Import date-fns for date formatting
import { format } from 'date-fns';

// Interface for data records (flexible to accommodate different data structures)
interface DataRecord {
  [key: string]: any;
}

// Reactive variables for managing component state and data
const proxyData = ref<DataRecord[]>([]); // Stores the fetched data
const availableFields = ref<string[]>([]); // Stores the available fields from the fetched data
const selectedFields = ref<{ [key: string]: boolean }>({}); // Stores the fields selected by the user for display
const contextFilter = ref(''); // Stores the context filter value
const showLoading = ref(false); // Controls the visibility of the loading spinner
const showData = ref(false); // Controls the visibility of the data table section

// Error handling variables
const showError = ref(false); // Controls the visibility of the error message section
const errorMessage = ref(''); // Stores the error message

// API response variable
const apiResponse = ref(''); // Stores the raw API response (optional)
const showApiResponse = ref(false); // Controls the visibility of the API response section (optional)

// Datepicker value (in Unix timestamp format)
const selectedDate = ref(Date.now()); // Stores the selected date as a Unix timestamp

// Sorting state management
const sortField = ref<string | null>(null); // Stores the field to sort by
const sortDirection = ref<'asc' | 'desc'>('asc'); // Stores the sort direction ('asc' or 'desc')
const collapseExpanded = ref(['select-fields']); // Controls the expansion state of the collapse panel

// Data Source Selection: Stores the currently selected data source (Litmus or Proxy)
const selectedDataSource = ref('proxy/litmus_data	'); // Default to proxy/litmus_data

// Watch for changes in collapse panel expansion
watch(collapseExpanded, (value) => {
  // Ensure that only the "Select Fields" panel can be expanded at a time
  if (value[0] === 'select-fields') {
    collapseExpanded.value = ['select-fields'];
  } else {
    collapseExpanded.value = [];
  }
});

// Computed Properties for Data Transformations

// Filtered data based on selected fields
const filteredData = computed(() => {
  // If no fields are selected, return all data
  if (Object.values(selectedFields.value).every((val) => !val)) {
    return proxyData.value;
  } else {
    // Otherwise, filter the data to include only selected fields
    return proxyData.value.map((record) => {
      const newRecord: DataRecord = {};
      for (const field in selectedFields.value) {
        if (selectedFields.value[field]) {
          newRecord[field] = record[field];
        }
      }
      return newRecord;
    });
  }
});

// List of visible fields based on user selection
const visibleFields = computed(() => {
  return Object.keys(selectedFields.value).filter((field) => selectedFields.value[field]);
});

// Sorted data based on selected sorting field and direction
const sortedData = computed(() => {
  if (!sortField.value) {
    return filteredData.value;
  }

  const data = [...filteredData.value];
  // Sort data based on the type of the field (numerical or string)
  data.sort((a, b) => {
    const aVal = a[sortField.value!];
    const bVal = b[sortField.value!];
    if (typeof aVal === 'number' && typeof bVal === 'number') {
      return sortDirection.value === 'asc' ? aVal - bVal : bVal - aVal; // Sort numbers numerically
    } else {
      return sortDirection.value === 'asc' ? String(aVal).localeCompare(String(bVal)) : String(bVal).localeCompare(String(aVal)); // Sort strings alphabetically
    }
  });

  return data;
});

// Data Fetching and Manipulation Functions

/**
 * Fetches data from the selected API endpoint (Litmus or Proxy) based on context and date filters.
 */
const fetchData = () => {
  showLoading.value = true;
  showData.value = false;
  showError.value = false; // Reset error state
  const options = {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' }
  };

  // Convert selectedDate (Unix timestamp) to YYYYMMDD format for the API call
  const formattedDate = format(new Date(selectedDate.value), 'yyyyMMdd');
  // Construct API URL based on selected data source
  const apiUrl = `/${selectedDataSource.value}?flatten=true&date=${formattedDate}&context=${contextFilter.value}`;

  fetch(apiUrl, options)
    .then((response) => {
      if (!response.ok) {
        // Assuming the error is in JSON format
        return response.json().then((errorJson) => {
          throw new Error(errorJson.error); // Use the error message from the JSON
        });
      }
      return response.json();
    })
    .then((data) => {
      proxyData.value = data;
      showData.value = true;
      availableFields.value = Object.keys(data[0]);

      // Store and display API response (optional)
      apiResponse.value = JSON.stringify(data, null, 2);

      // Select important fields by default
      selectedFields.value = availableFields.value.reduce(
        (acc, field) => ({
          ...acc,
          [field]: isImportantField(field)
        }),
        {}
      );

      showLoading.value = false;
    })
    .catch((err) => {
      console.error(err);
      showLoading.value = false;
      showError.value = true;
      errorMessage.value = err.message.replace(/\n/g, '<br>'); // Replace newlines with <br> for HTML display
    });
};

/**
 * Checks if a field is considered important for default selection.
 * @param {string} field - The name of the field.
 * @returns {boolean} - True if the field is important, otherwise false.
 */
const isImportantField = (field: string): boolean => {
  // Define keywords that indicate important fields
  const keywords = ['_text', 'timestamp', 'totaltokencount'];
  return keywords.some((keyword) => field.toLowerCase().includes(keyword));
};

// Data Table Sorting Function

/**
 * Sorts the data table by the given field.
 * @param {string} field - The name of the field to sort by.
 */
const sortTable = (field: string) => {
  // Toggle sort direction if the same field is clicked again
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc';
  } else {
    // Set the new sort field and default to ascending order
    sortField.value = field;
    sortDirection.value = 'asc';
  }
};

// Data Export Function

/**
 * Exports data to a CSV file.
 */
const exportCSV = () => {
  // Convert data to CSV format
  const csvContent = convertToCSV(proxyData.value);
  // Download the CSV file with a filename based on the selected data source
  downloadFile(csvContent, `${selectedDataSource.value}.csv`, 'text/csv');
};

/**
 * Converts an array of objects to CSV format.
 * @param {any[]} data - The array of objects to convert.
 * @returns {string} - The CSV string.
 */
function convertToCSV(data: any[]): string {
  if (!data || data.length === 0) {
    return '';
  }

  const headers = Object.keys(data[0]);
  let csvString = headers.join(',') + '\n';

  data.forEach((row) => {
    const rowValues = headers.map((header) => {
      const value = row[header];
      return typeof value === 'string' ? `"${value.replace(/"/g, '""')}"` : value; // Escape double quotes within strings
    });
    csvString += rowValues.join(',') + '\n';
  });

  return csvString;
}

/**
 * Downloads a file with the given content and filename.
 * @param {string} content - The content of the file.
 * @param {string} filename - The desired filename.
 * @param {string} mimeType - The MIME type of the file.
 */
function downloadFile(content: string, filename: string, mimeType: string) {
  const blob = new Blob([content], { type: mimeType });
  const url = window.URL.createObjectURL(blob);

  // Create a temporary link element to trigger the download
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  a.click();

  // Clean up the URL object
  window.URL.revokeObjectURL(url);
}

// Lifecycle Hooks

// Fetch data when the component is mounted
onMounted(() => {
  fetchData();
});
</script>

<style>
/* CSS Styles for the Component */
.wrap-text {
  white-space: pre-wrap;
  word-break: break-all;
  min-width: 20em;
}

.filter-container {
  display: flex;
  gap: 10px;
  margin-bottom: 10px;
  align-items: center; /* Align items vertically in the container */
}

td {
  align-content: baseline;
}

.checkbox-container {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 10px;
}

.sortable-header {
  cursor: pointer;
}

.export-buttons {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 10px;
}

.error-message {
  display: flex;
  align-items: center;
  gap: 10px;
  color: red;
  margin-bottom: 10px;
}

.api-response {
  margin-top: 20px;
  border: 1px solid #ccc;
  padding: 10px;
  background-color: #f8f8f8;
  white-space: pre-wrap; /* Allow wrapping for long responses */
}

.table-container {
  overflow-x: auto; /* Enable horizontal scrolling */
}

.litmus-data-selector {
  width: 30%;
}
</style>
