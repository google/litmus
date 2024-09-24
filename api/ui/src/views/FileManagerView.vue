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
  <!-- Breadcrumbs for navigation -->
  <Breadcrumbs :title="page.title"></Breadcrumbs>

  <!-- Page container -->
  <div class="page">
    <!-- Container for the components list -->
    <div class="components-list">
      <!-- Scrollbar for the Run component -->
      <n-scrollbar x-scrollable style="width: 100%">
        <div v-if="show">
          <n-spin :show="show">
            <div>Loading...</div>
          </n-spin>
        </div>
        <div v-else>
          <n-upload
            ref="uploadRef"
            multiple
            directory-dnd
            :default-upload="false"
            :on-change="handleChange"
            :on-finish="handleFinish"
            :on-error="handleError"
          >
            <n-upload-dragger>
              <div style="margin-bottom: 12px">
                <n-icon size="48" :depth="3">
                  <ArchiveIcon />
                </n-icon>
              </div>
              <n-text style="font-size: 16px"> Click or drag a file to this area to upload </n-text>
            </n-upload-dragger>
          </n-upload>
          <v-btn variant="flat" color="primary" class="mt-2" @click="handleUpload">Upload</v-btn>
          <n-divider />
          <n-space>
            <v-row class="page-breadcrumb mb-0 mt-n2">
              <v-col cols="12" md="12">
                <v-card elevation="0" variant="text">
                  <v-row no-gutters class="align-center">
                    <v-col sm="12">
                      <h3 class="text-h3 mt-1 mb-0">Uploaded Files</h3>
                    </v-col>
                  </v-row>
                </v-card>
              </v-col>
            </v-row>
          </n-space>
          <n-list bordered>
            <n-list-item v-for="file in files" :key="file.name">
              <n-space>
                <n-tag type="info" size="small" @click="copyReference(file.name)"> [FILE: {{ file.name }}] </n-tag>
                <v-btn :href="'/files/' + file.name" target="_blank" size="small">Download</v-btn>
                <v-btn size="small" type="error" variant="flat" color="accent" @click="deleteFile(file.name)">Delete</v-btn>
              </n-space>
            </n-list-item>
          </n-list>
        </div>
      </n-scrollbar>
    </div>
  </div>
</template>

<script lang="ts" setup>
// Import necessary components and libraries
import {
  NScrollbar,
  NUpload,
  NUploadDragger,
  NIcon,
  NText,
  NSpace,
  NButton,
  NList,
  NListItem,
  NThing,
  NTag,
  NDivider,
  NSpin,
  NCard,
  useMessage
} from 'naive-ui';
import { ref, onMounted } from 'vue';
import Breadcrumbs from '@/components/shared/Breadcrumbs.vue';
import { ArchiveOutline as ArchiveIcon } from '@vicons/ionicons5';

// Define a reactive object for the page title
const page = ref({ title: 'File Upload' });
const uploadRef = ref<any>(null);
const message = useMessage();
const files = ref<FileObject[]>([]);
const fileList = ref<any[]>([]);
const show = ref(false); // Loading state

// Define a type for file objects
interface FileObject {
  name: string;
  gcs_path: string;
}

// Fetch uploaded files on component mount
onMounted(async () => {
  show.value = true;
  await fetchFiles();
  show.value = false;
});

// Handle file selection changes
const handleChange = (file: any) => {
  fileList.value = file.fileList;
};

// Handle upload finish
const handleFinish = ({ file }: { file: any }) => {
  message.success(`File ${file.name} uploaded successfully`);
  fetchFiles(); // Refresh file list after upload
};

// Handle upload errors
const handleError = ({ file }: { file: any }) => {
  message.error(`Failed to upload file ${file.name}`);
};

// Trigger manual upload of files
const handleUpload = () => {
  fileList.value.forEach((file) => {
    if (file.status === 'pending') {
      // Create a FormData object to send the file
      const formData = new FormData();
      formData.append('file', file.file);

      // Send a POST request to upload the file
      fetch(`/files/${file.name}`, {
        method: 'POST',
        body: formData
      })
        .then((response) => {
          if (!response.ok) {
            throw new Error('File upload failed');
          }
          return response.json();
        })
        .then((data) => {
          // Handle successful upload (e.g., update file status)
          file.status = 'finished';
          message.success(data.message);
          fetchFiles(); // Refresh file list after upload
        })
        .catch((error) => {
          // Handle upload error (e.g., update file status, show error message)
          file.status = 'error';
          message.error(error.message);
        });
    }
  });
};

/**
 * Deletes a file from the files bucket.
 * @param {string} filename - The name of the file to delete.
 */
const deleteFile = async (filename: string) => {
  try {
    // Send a DELETE request to the backend to delete the file
    const response = await fetch(`/files/${filename}`, {
      method: 'DELETE'
    });

    if (response.ok) {
      // If the deletion was successful, update the file list
      await fetchFiles();
      // Show a success message
      message.success(`File ${filename} deleted successfully`);
    } else {
      // If there was an error, parse the error data from the response
      const errorData = await response.json();
      // Show an error message with the error details
      message.error(`Error deleting file: ${errorData.error}`);
    }
  } catch (error) {
    console.error('Error deleting file:', error);
    message.error('An unexpected error occurred while deleting the file.');
  }
};

// Fetch the list of files from the backend.
const fetchFiles = async () => {
  try {
    const response = await fetch('/files');
    const data = await response.json();
    // Ensure data.files is an array of FileObject
    files.value = data.files as FileObject[];
  } catch (error) {
    console.error('Error fetching files:', error);
    message.error('Failed to fetch files');
  }
};

/**
 * Copies the provided reference to the clipboard.
 * @param {string} reference - The reference to copy.
 */
const copyReference = (reference: string) => {
  navigator.clipboard
    .writeText('[FILE: ' + reference + ']')
    .then(() => {
      message.success('Reference copied to clipboard!');
    })
    .catch((error) => {
      console.error('Failed to copy reference:', error);
      message.error('Failed to copy reference!');
    });
};
</script>
<style scoped lang="scss">
.n-tag {
  // Ensure the list takes up the full width and adjusts for columns
  cursor: pointer;
}
</style>
