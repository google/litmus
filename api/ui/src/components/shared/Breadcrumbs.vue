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
  <v-row class="page-breadcrumb mb-0 mt-n2">
    <v-col cols="12" md="12">
      <v-card elevation="0" variant="text">
        <v-row no-gutters class="align-center">
          <v-col sm="12">
            <v-breadcrumbs
              :items="props.breadcrumbs"
              class="text-h6 pa-1 font-weight-medium mb-0"
            >
              <template v-slot:divider>
                <div class="d-flex align-center">/</div>
              </template>
              <template v-slot:prepend>
                <router-link to="/" class="text-lightText text-h6 text-decoration-none">
                  Home
                </router-link>
                <div class="d-flex align-center px-2">/</div>
              </template>
            </v-breadcrumbs>
            <h3 class="text-h3 mt-1 mb-0">{{ props.title }}</h3>
          </v-col>
        </v-row>
      </v-card>
    </v-col>
  </v-row>
</template>
<script setup lang="ts">
const props = defineProps({
  title: String,
  /* eslint-disable @typescript-eslint/no-explicit-any */
  breadcrumbs: Array as any,
  icon: String,
});
</script>
<style lang="scss">
.page-breadcrumb {
  .v-toolbar {
    background: transparent;
  }
}
</style>
