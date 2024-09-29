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
    <!-- Page Title and Breadcrumb -->
    <v-row class="page-breadcrumb mb-0 mt-n2">
      <v-col cols="12" md="12">
        <v-card elevation="0" variant="text">
          <v-row no-gutters class="align-center">
            <v-col sm="12">
              <h3 class="text-h3 mt-1 mb-0">Run: {{ runId }} (Template: {{ templateId }})</h3>
            </v-col>
          </v-row>
        </v-card>
      </v-col>
    </v-row>

    <!-- Filter and Action Buttons -->
    <div class="filter-container">
      <n-input v-model:value="requestFilter" placeholder="Request Filter (e.g., body.query)" clearable @change="fetchRunDetails" />
      <n-input
        v-model:value="responseFilter"
        placeholder="Response Filter (e.g., response.output.text,assessment,status)"
        clearable
        @change="fetchRunDetails"
      />
      <!-- Show Only Flagged Toggle -->
      <n-switch v-model:value="showOnlyFlagged" @update:value="fetchRunDetails" />
      <span><v-icon>mdi-flag</v-icon></span>
      <!-- Clear Filters Button -->
      <v-btn variant="flat" color="primary" @click="clearFilter"> Clear Filters </v-btn>
      <!-- Export Test Cases Button -->
      <v-btn variant="flat" color="primary" @click="exportTestCasesCSV"> Export Test Cases to CSV </v-btn>
    </div>

    <!-- Test Cases Table -->
    <n-spin :show="show">
      <n-table :data="testCases" class="table-min-width" striped>
        <thead>
          <tr>
            <th>Test Case</th>
            <th>Input</th>
            <th>Output</th>
          </tr>
        </thead>
        <tbody>
          <tr v-if="filteredTestCases.length === 0">
            <td colspan="3">No Results</td>
          </tr>
          <tr v-for="testCase in filteredTestCases" :key="testCase.id">
            <td>
              <div class="flag-container">
                <v-btn icon variant="text" size="small" @click="toggleFlagTestCase(testCase.id)">
                  <v-icon v-if="testCase.flagged">mdi-flag</v-icon>
                  <v-icon v-else>mdi-flag-outline</v-icon>
                </v-btn>
                <strong>{{ testCase.id }}</strong>
              </div>
              <n-divider v-if="testCase.response.status"></n-divider>
              <n-space justify="center" v-if="testCase.response.status" size="small">
                <span v-if="testCase.response.status === 'Failed'" style="color: red">
                  <Icon :name="FailedIcon" :size="30" />
                </span>
                <span v-if="testCase.response.status === 'Passed'" style="color: green">
                  <Icon :name="SuccessIcon" :size="30" />
                </span>
              </n-space>
              <n-divider></n-divider>
              {{ testCase.tracing_id }}
              <n-divider></n-divider>
              <n-space justify="center" size="small">
                <div class="rating-container">
                  <n-rate v-model:value="testCase.rating" @update:value="rateTestCase(testCase.id)" />
                </div>
              </n-space>
              <n-divider></n-divider>
              <n-flex vertical>
                <v-btn
                  size="small"
                  prepend-icon="mdi-magnify-expand"
                  variant="outlined"
                  color="primary"
                  @click="openDrawer(testCase.tracing_id)"
                >
                  Explore
                </v-btn>
                <!-- Comments Button -->
                <v-btn
                  prepend-icon="mdi-comment-plus"
                  variant="outlined"
                  color="primary"
                  class="mt-2"
                  @click="openCommentsDrawer(testCase.id)"
                  size="small"
                >
                  Comment ({{ testCase.comment_count }})
                </v-btn>
              </n-flex>
            </td>
            <!-- Input Data -->
            <td>
              <strong>Request:</strong>
              <pre class="wrap-text">
                {{ JSON.stringify(testCase.request, null, 2) }}
              </pre>
              <n-divider></n-divider>
              <strong>Golden Answer:</strong>
              <pre class="wrap-text">
                {{ JSON.stringify(testCase.golden_response, null, 2) }}
              </pre>
            </td>
            <!-- Output Data -->
            <td>
              <pre class="wrap-text" v-if="testCase.response.error">
                {{ testCase.response.error }}
              </pre>
              <pre class="wrap-text" v-else>
                {{ JSON.stringify(testCase.response, null, 2) }}
              </pre>
            </td>
          </tr>
        </tbody>
      </n-table>
    </n-spin>

    <!-- Data Drawer -->
    <n-drawer v-model:show="showDrawer" :width="980">
      <n-drawer-content :title="drawerTitle" :native-scrollbar="false" :width="996" closable>
        <n-spin :show="showDrawerSpinner">
          <div v-if="drawerError" class="error-message">
            <n-icon size="30" color="red">
              <AlertCircle />
            </n-icon>
            <p>
              Error: There is no data available in BigQuery for
              {{ drawerTitle }} on {{ selectedDate }}. This might be due to a temporary issue or the data might not be available.
            </p>
          </div>
          <div v-else>
            <!-- Test Info Section: Displays proxy/agg data when drawerContent is available -->
            <div v-if="drawerContent.length > 0" class="test-info">
              <strong>Total Tokens:</strong> {{ testInfo.total_token_count }} <br />
              <strong>Prompt Tokens:</strong> {{ testInfo.prompt_token_count }} <br />
              <strong>Candidate Tokens:</strong> {{ testInfo.candidates_token_count }} <br />
              <strong>Latency (ms):</strong> {{ testInfo.average_latency }}
            </div>

            <!-- Field Selection Collapse Panel -->
            <n-collapse v-model:expanded="collapseExpanded">
              <n-collapse-item title="Select Fields" name="select-fields">
                <div class="checkbox-container">
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

            <!-- Export Buttons -->
            <div class="export-buttons">
              <v-btn variant="flat" color="primary" @click="exportCSV"> Export Table to CSV </v-btn>
            </div>

            <!-- Data Table -->
            <div class="table-container">
              <n-table class="table-min-width" striped>
                <thead>
                  <tr>
                    <th v-for="field in visibleFields" :key="field" @click="sortTable(field)" class="sortable-header">
                      {{ field }}
                      <span v-if="sortField === field">
                        <n-icon :component="sortDirection === 'asc' ? CaretUpOutline : CaretDownOutline" />
                      </span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(record, index) in sortedData" :key="index">
                    <td v-for="field in visibleFields" :key="field">
                      <template v-if="typeof record[field] === 'object'">
                        {{ JSON.stringify(record[field], null, 2) }}
                      </template>
                      <template v-else>
                        {{ record[field] }}
                      </template>
                    </td>
                  </tr>
                </tbody>
              </n-table>
            </div>
          </div>
          <n-space v-if="!drawerError">
            <v-btn variant="flat" color="primary" @click="exportJSON"> Export all Data to JSON File </v-btn>
          </n-space>
        </n-spin>
      </n-drawer-content>
    </n-drawer>

    <!-- Comments Drawer -->
    <n-drawer v-model:show="showCommentsDrawer" placement="left" :width="500">
      <n-drawer-content :title="'Comments for ' + currentTestCaseId" closable>
        <!-- Comment List -->
        <n-list bordered>
          <n-list-item v-for="(comment, index) in currentTestCaseComments" :key="index">
            {{ comment }}
          </n-list-item>
        </n-list>
        <!-- Comment Input -->
        <n-input
          v-model:value="newComment"
          type="textarea"
          placeholder="Add a comment"
          :autosize="{
            minRows: 2
          }"
        />
        <n-space>
          <!-- Save Comment Button -->
          <v-btn variant="flat" color="primary" class="mt-2" @click="addComment(currentTestCaseId)">Add Comment</v-btn>
        </n-space>
      </n-drawer-content>
    </n-drawer>
  </div>
</template>

<script lang="ts" setup>
// Import necessary components from UI libraries
import {
  NTable,
  NInput,
  NSpin,
  NDivider,
  NDrawer,
  NDrawerContent,
  NCheckbox,
  NIcon,
  NCollapse,
  NCollapseItem,
  NSpace,
  useMessage,
  NSwitch,
  NRate,
  NList,
  NListItem,
  NFlex
} from 'naive-ui';
// Import routing functionality from Vue Router
import { useRoute } from 'vue-router';
// Import reactivity functions from Vue
import { ref, onMounted, computed, watch } from 'vue';
// Import icons from icon libraries
import { CaretUpOutline, CaretDownOutline } from '@vicons/ionicons5';
import { AlertCircle } from '@vicons/ionicons5';
// Import custom Icon component
import Icon from '@/components/common/Icon.vue';

// Define names for success and failure icons
const SuccessIcon = 'carbon:checkmark-outline';
const FailedIcon = 'carbon:close-filled';

// Interface for Test Case Data Structure
interface TestCase {
  id: string;
  response: {
    status: string;
    error?: string;
  };
  request: any;
  golden_response: any;
  tracing_id: string;
  rating: number;
  flagged: boolean;
  comments: string[];
  comment_count: number;
}

// Interface for Generic Data Records
interface DataRecord {
  [key: string]: any;
}

// --- Data and State Management ---

// Get Route Parameters
const route = useRoute();
// Get the runId from the route parameters
const runId = route.params.runId as string;

// Reactive Data for Test Cases, Drawer, Loading State, Date, Filters, etc.
// Test case data fetched from the API
const testCases = ref<TestCase[]>([]);
// Controls the visibility of the data drawer
const showDrawer = ref(false);
// Title of the data drawer
const drawerTitle = ref('');
// Content displayed in the data drawer
const drawerContent = ref<DataRecord[]>([]);
// Flag to indicate if an error occurred while fetching drawer content
const drawerError = ref(false);
// Flag to control the loading spinner for the test cases table
const show = ref(false);
// Flag to control the loading spinner for the data drawer
const showDrawerSpinner = ref(false);
// Flag to control the visibility of the comments drawer
const showCommentsDrawer = ref(false);
// Current test case ID for the comments drawer
const currentTestCaseId = ref('');

// Get Current Date for Default Selection
const today = new Date();
const year = today.getFullYear();
const month = String(today.getMonth() + 1).padStart(2, '0');
const day = String(today.getDate()).padStart(2, '0');
// Selected date for filtering data in the drawer
const selectedDate = ref(`${year}${month}${day}`);

// Filters for Request, Response, and Golden Responses
// Filter for response data
const responseFilter = ref('assessment,response.output.text,response.output.intent,status');
// Filter for request data
const requestFilter = ref('body.query');
// ID of the template used for the test run
const templateId = ref('');
// Filter for golden response data
const goldenResponsesFilter = ref('');
// Filter to show only flagged runs
const showOnlyFlagged = ref(false);

// Data Table Field Management
// List of available fields in the drawer data
const availableFields = ref<string[]>([]);
// Keeps track of which fields are currently selected for display
const selectedFields = ref<{ [key: string]: boolean }>({});
// The field by which the data table is currently sorted
const sortField = ref<string | null>(null);
// The direction of sorting (ascending or descending)
const sortDirection = ref<'asc' | 'desc'>('asc');
// Controls the expansion state of collapse panels in the drawer
const collapseExpanded = ref(['select-fields']);

// New comment for a test case
const newComment = ref('');

// Reactive object to store test info from the API
const testInfo = ref<Record<string, any>>({});

// Get the message instance from Naive UI
const message = useMessage(); // Import useMessage

// Watch for Changes in Collapse Panel Expansion
watch(collapseExpanded, (value) => {
  // Ensure that only the 'select-fields' panel can be expanded at a time
  if (value[0] === 'select-fields') {
    collapseExpanded.value = ['select-fields'];
  } else {
    collapseExpanded.value = [];
  }
});

// Watch for changes in drawerContent
watch(
  drawerContent,
  (newDrawerContent) => {
    if (newDrawerContent.length > 0) {
      showTestInfo(drawerTitle.value);
    } else {
      // Reset test info when drawerContent is empty
      testInfo.value = {};
    }
  },
  { deep: true }
);

// --- Computed Properties for Data Transformations ---

// Filter Data Based on Selected Fields
const filteredData = computed(() => {
  // If no fields are selected, return the original data
  if (Object.values(selectedFields.value).every((val) => !val)) {
    return drawerContent.value;
  } else {
    // Otherwise, filter the data to include only selected fields
    return drawerContent.value.map((record) => {
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

// Get a List of Visible Fields (i.e., those that are selected)
const visibleFields = computed(() => {
  return Object.keys(selectedFields.value).filter((field) => selectedFields.value[field]);
});

// Sort the Filtered Data Based on the Selected Sorting Field and Direction
const sortedData = computed(() => {
  // If no sorting field is selected, return the filtered data as is
  if (!sortField.value) {
    return filteredData.value;
  }

  const data = [...filteredData.value];

  // Sort the data based on the selected field and direction
  data.sort((a, b) => {
    const aVal = a[sortField.value!]; // Non-null assertion here
    const bVal = b[sortField.value!]; // Non-null assertion here

    // Handle sorting for different data types
    if (typeof aVal === 'number' && typeof bVal === 'number') {
      // Numerical sorting
      return sortDirection.value === 'asc' ? aVal - bVal : bVal - aVal;
    } else {
      // String sorting
      return sortDirection.value === 'asc' ? String(aVal).localeCompare(String(bVal)) : String(bVal).localeCompare(String(aVal));
    }
  });

  return data;
});

// Computed property for filtering test cases by flagged status
const filteredTestCases = computed(() => {
  if (showOnlyFlagged.value) {
    return testCases.value.filter((testCase) => testCase.flagged);
  } else {
    return testCases.value;
  }
});

// Computed property to retrieve comments for the current test case
const currentTestCaseComments = computed(() => {
  const testCase = testCases.value.find((tc) => tc.id === currentTestCaseId.value);
  return testCase ? testCase.comments : [];
});

// --- Data Fetching and Manipulation Functions ---

/**
 * Converts a date string to YYYYMMDD format.
 * @param {string} dateString - The date string to convert.
 * @returns {string} - The date in YYYYMMDD format.
 */
function convertDate(dateString: string): string {
  const date = new Date(dateString);
  const year = date.getFullYear();
  const month = String(date.getMonth() + 1).padStart(2, '0');
  const day = String(date.getDate()).padStart(2, '0');
  return `${year}${month}${day}`;
}

/**
 * Fetches the run details from the API based on applied filters.
 */
const fetchRunDetails = () => {
  show.value = true; // Show loading spinner
  // Construct the filter string for the API request
  let filterString = `response_filter=${responseFilter.value}&request_filter=${requestFilter.value}&golden_responses_filter=${goldenResponsesFilter.value}`;

  // Add flagged filter if enabled
  if (showOnlyFlagged.value) {
    filterString += '&flagged=true';
  }

  // Fetch run details from the API
  fetch(`/runs/status/${runId}?${filterString}`)
    .then((response) => response.json())
    .then((data) => {
      // Update testCases with fetched data, adding user interaction details
      testCases.value = (data.testCases as TestCase[]).map((testCase) => ({
        ...testCase
      }));
      show.value = false; // Hide loading spinner after data is fetched
    })
    .catch((error) => {
      console.error('Error fetching run details:', error);
      show.value = false; // Hide loading spinner on error
    });
};

/**
 * Clears all applied filters and re-fetches the run details.
 */
const clearFilter = () => {
  responseFilter.value = 'response'; // Reset response filter
  requestFilter.value = ''; // Reset request filter
  goldenResponsesFilter.value = ''; // Reset golden response filter
  showOnlyFlagged.value = false; // Reset flagged filter
  fetchRunDetails(); // Refetch data with cleared filters
};

/**
 * Opens the data drawer and fetches data for the specified trace ID.
 * @param {string} trace_id - The ID of the trace to fetch data for.
 */
const openDrawer = (trace_id: string) => {
  showDrawerSpinner.value = true; // Show drawer loading spinner
  drawerTitle.value = trace_id; // Set drawer title
  drawerContent.value = []; // Clear previous drawer content
  drawerError.value = false; // Reset drawer error state

  const options = {
    method: 'GET',
    headers: { 'Content-Type': 'application/json' }
  };

  // Fetch data from the API for the specified trace ID and date
  fetch(`/proxy/data?flatten=true&date=${selectedDate.value}&context=${trace_id}`, options)
    .then((response) => {
      // Handle BigQuery data errors
      if (!response.ok && Math.floor(response.status / 100) === 5) {
        throw new Error('BigQuery Data Error');
      }
      return response.json();
    })
    .then((data: DataRecord[]) => {
      // Extract available fields from the fetched data
      availableFields.value = Object.keys(data[0]);

      // Set default selected fields based on importance
      selectedFields.value = availableFields.value.reduce(
        (acc, field) => ({
          ...acc,
          [field]: isImportantField(field)
        }),
        {}
      );

      // Update drawer content and hide loading spinner
      drawerContent.value = data;
      showDrawerSpinner.value = false;

      // Fetch test info after the table data is loaded
      if (drawerContent.value.length > 0) {
        showTestInfo(drawerTitle.value);
      }
    })
    .catch((err) => {
      console.error(err);
      showDrawerSpinner.value = false;
      drawerError.value = true; // Show error message in the drawer
    });

  showDrawer.value = true; // Open the drawer
};

/**
 * Fetches run fields (date, template ID, input/output fields) from the API.
 */
const fetchRunFields = () => {
  // Fetch run fields from the API
  fetch(`/runs/status_fields/${runId}`)
    .then((response) => response.json())
    .then((data) => {
      // Update relevant data properties based on fetched fields
      selectedDate.value = convertDate(data.run_date);
      templateId.value = data.template_id;
      requestFilter.value = data.template_input_field;
      responseFilter.value = 'assessment,response.' + data.template_output_field + ',status';
      fetchRunDetails(); // Fetch run details after updating filters
    })
    .catch((error) => {
      console.error('Error fetching run details:', error);
    });
};

/**
 * Fetches test information from the proxy/agg endpoint.
 * @param traceId The tracing ID of the test case.
 */
const showTestInfo = async (traceId: string) => {
  try {
    const date = convertDate(selectedDate.value);
    const response = await fetch(`/proxy/agg?date=${date}&context=${traceId}`);
    if (!response.ok) {
      throw new Error('Failed to fetch test info');
    }
    const data = await response.json();

    // Update testInfo with the fetched data
    if (data.length > 0) {
      testInfo.value = data[0];
    } else {
      testInfo.value = {};
    }
  } catch (error) {
    console.error('Error fetching test info:', error);
    message.error('Failed to fetch test info');
  }
};

// --- Data Table Sorting Function ---

/**
 * Sorts the data table by the specified field.
 * @param {string} field - The field to sort the table by.
 */
const sortTable = (field: string) => {
  // If the field is already the sorting field, toggle the sort direction
  if (sortField.value === field) {
    sortDirection.value = sortDirection.value === 'asc' ? 'desc' : 'asc';
  } else {
    // Otherwise, set the new sorting field and default to ascending order
    sortField.value = field;
    sortDirection.value = 'asc';
  }
};

// --- Helper Functions ---

/**
 * Checks if a field is considered important for default selection in the field selection panel.
 * @param {string} field - The field name to check.
 * @returns {boolean} - True if the field is important, otherwise false.
 */
const isImportantField = (field: string): boolean => {
  const keywords = ['_text', 'timestamp', 'totaltokencount'];
  return keywords.some((keyword) => field.toLowerCase().includes(keyword));
};

/**
 * Converts JSON data to CSV format.
 * @param {any[]} jsonData - The JSON data to convert.
 * @returns {string} - The CSV representation of the data.
 */
function jsonToCsv(jsonData: any[]): string {
  if (!jsonData || jsonData.length === 0) {
    return '';
  }

  // Extract headers from the first object
  const keys = Object.keys(jsonData[0]);

  // Construct the CSV content string
  const csvContent =
    keys.join(',') +
    '\n' +
    jsonData
      .map((row) => {
        return keys
          .map((key) => {
            let value = row[key];
            // Escape double quotes in string values
            if (typeof value === 'string') {
              value = value.replace(/"/g, '""');
              return '"' + value + '"';
            } else {
              return value;
            }
          })
          .join(',');
      })
      .join('\n');

  return csvContent;
}

/**
 * Downloads a file with the given content, filename, and MIME type.
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

/**
 * Exports the currently visible table data to a CSV file.
 */
const exportCSV = () => {
  const csvContent = jsonToCsv(sortedData.value);
  downloadFile(csvContent, 'exported_data.csv', 'text/csv');
};

/**
 * Exports all data in the drawer to a JSON file.
 */
const exportJSON = () => {
  const jsonContent = JSON.stringify(drawerContent.value);
  downloadFile(jsonContent, 'exported_data.json', 'application/json');
};

/**
 * Extracts data from an object based on a dot-separated path.
 * @param {any} obj - The object to extract data from.
 * @param {string} path - The dot-separated path to the desired data.
 * @returns {any} - The data at the specified path.
 */
const extractDataByPath = (obj: any, path: string): any => {
  return path.split('.').reduce((o, k) => (o && o[k] ? o[k] : ''), obj);
};

/**
 * Prepares test case data for CSV export.
 * @param {TestCase[]} data - The test case data to prepare.
 * @returns {any[]} - The prepared data for CSV export.
 */
const prepareTestCasesForCSV = (data: TestCase[]): any[] => {
  return data.map((testCase) => {
    const csvRow: { [key: string]: any } = {};

    // Mapping of CSV headers to data paths
    const headerMapping: { [key: string]: string } = {
      'Test Case ID': 'id',
      'Tracing ID': 'tracing_id',
      Status: 'response.status',
      Request: 'request',
      'Golden Response': 'golden_response',
      Response: 'response',
      Rating: 'rating',
      Flagged: 'flagged',
      Comments: 'comments'
    };

    // Extract data for each header
    for (const header in headerMapping) {
      const dataPath = headerMapping[header];
      csvRow[header] = extractDataByPath(testCase, dataPath);
      // Stringify object values for CSV export
      if (typeof csvRow[header] === 'object') {
        csvRow[header] = JSON.stringify(csvRow[header]);
      }
    }

    return csvRow;
  });
};

/**
 * Exports all test cases data to a CSV file.
 */
const exportTestCasesCSV = () => {
  const csvContent = jsonToCsv(prepareTestCasesForCSV(testCases.value));
  downloadFile(csvContent, 'test_cases.csv', 'text/csv');
};

/**
 * Rates a test case and refreshes the rating from the API.
 * @param testCaseId The ID of the test case to rate.
 */
const rateTestCase = async (testCaseId: string) => {
  try {
    // Get the new rating value from the corresponding test case
    const testCase = testCases.value.find((tc) => tc.id === testCaseId);
    const newRating = testCase ? testCase.rating : 0; // Default to 0 if test case not found

    const response = await fetch(`/runs/${runId}/${testCaseId}/rate`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ rating: newRating })
    });

    if (response.ok) {
      message.success('Test case rated successfully!');
      // After a successful rating update, re-fetch the run details
      fetchRunDetails();
    } else {
      const errorData = await response.json();
      message.error(`Error rating test case: ${errorData.error}`);
    }
  } catch (error) {
    console.error('Error rating test case:', error);
    message.error('An unexpected error occurred while rating the test case.');
  }
};

/**
 * Toggles the flagged status of a test case.
 * @param testCaseId The ID of the test case to toggle.
 */
const toggleFlagTestCase = async (testCaseId: string) => {
  try {
    const response = await fetch(`/runs/${runId}/${testCaseId}/flag`, {
      method: 'PUT',
      headers: {
        'Content-Type': 'application/json'
      }
    });

    if (response.ok) {
      // Find the test case and update its flagged status
      const testCase = testCases.value.find((tc) => tc.id === testCaseId);
      if (testCase) {
        testCase.flagged = !testCase.flagged;
      }
      message.success('Test case flagged status updated!');
    } else {
      const errorData = await response.json();
      message.error(`Error updating flagged status: ${errorData.error}`);
    }
  } catch (error) {
    console.error('Error updating flagged status:', error);
    message.error('An unexpected error occurred while updating the flagged status.');
  }
};

/**
 * Opens the comments drawer for a specific test case.
 * @param testCaseId The ID of the test case.
 */
const openCommentsDrawer = (testCaseId: string) => {
  // Set the current test case ID
  currentTestCaseId.value = testCaseId;
  // Open the drawer
  showCommentsDrawer.value = true;
  // Clear the new comment input field
  newComment.value = '';
};

/**
 * Adds a new comment to the specified test case.
 * @param testCaseId The ID of the test case to add the comment to.
 */
const addComment = async (testCaseId: string) => {
  try {
    // Check if the new comment is empty
    if (newComment.value.trim() === '') {
      message.warning('Please enter a comment.');
      return;
    }

    const response = await fetch(`/runs/${runId}/${testCaseId}/comment`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ comment: newComment.value })
    });

    if (response.ok) {
      // Find the test case and add the new comment
      const testCase = testCases.value.find((tc) => tc.id === testCaseId);
      if (testCase) {
        // Initialize comments array if it's null
        if (!testCase.comments) {
          testCase.comments = [];
        }
        // Push the new comment
        testCase.comments.push(newComment.value);
        // Increase comment count
        testCase.comment_count = testCase.comment_count + 1;
        newComment.value = ''; // Clear the comment input field
      }
      message.success('Comment added successfully!');
      // Re-fetch run details to get updated comments
      fetchRunDetails();
    } else {
      const errorData = await response.json();
      message.error(`Error adding comment: ${errorData.error}`);
    }
  } catch (error) {
    console.error('Error adding comment:', error);
    message.error('An unexpected error occurred while adding the comment.');
  }
};

// --- Lifecycle Hook ---

// Fetch Run Fields and Initial Run Details When Component is Mounted
onMounted(() => {
  fetchRunFields();
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
  align-items: center;
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
}

.table-container {
  overflow-x: auto; /* Enable horizontal scrolling */
}

/* Styles for rating and flag containers */
.rating-container,
.flag-container {
  display: flex;
  align-items: center;
  gap: 5px;
  margin-top: 5px;
}

.comments-section {
  margin-top: 10px;
}

.comment {
  border: 1px solid #ccc;
  padding: 10px;
  margin-bottom: 5px;
  border-radius: 4px;
}

.test-info {
  margin-bottom: 15px; /* Add margin between test info and table */
}
</style>
