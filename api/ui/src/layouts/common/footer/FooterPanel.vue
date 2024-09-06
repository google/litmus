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
  <v-footer class="px-0 footer">
    <v-row justify="center" no-gutters>
      <v-col cols="6">
        <p class="text-caption mb-0">{{ version }}</p>
      </v-col>
      <v-col class="text-right" cols="6">
        <a v-for="(item, i) in footerLink" :key="i" class="mx-2 text-caption text-darkText" :href="item.url">
          {{ item.title }}
        </a>
      </v-col>
    </v-row>
  </v-footer>
</template>
<script setup lang="ts">
import { ref, onMounted, shallowRef } from 'vue';

const footerLink = shallowRef([
  {
    title: 'Github',
    url: 'https://github.com/google/litmus'
  },
  {
    title: 'Privacy',
    url: 'https://cloud.google.com/terms/cloud-privacy-notice'
  },
  {
    title: 'Terms',
    url: 'https://cloud.google.com/terms'
  }
]);

const version = ref(null);

onMounted(() => {
  get_version();
});

async function get_version() {
  const requestOptions = {
    method: 'GET'
  };

  const response = await fetch('/version', requestOptions);
  const data = await response.json();

  version.value = data.version;
}
</script>
