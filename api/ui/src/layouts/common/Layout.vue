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
  <n-message-provider>
    <v-locale-provider>
      <v-app :class="[]">
        <SidebarVue />
        <HeaderVue />
        <v-main class="page-wrapper">
          <v-container>
            <div>
              <!-- Loader start -->
              <LoaderWrapper />
              <!-- Loader end -->
              <RouterView />
            </div>
          </v-container>
          <v-container class="pt-0">
            <div>
              <FooterPanel />
            </div>
          </v-container>
        </v-main>
      </v-app>
    </v-locale-provider>
  </n-message-provider>
</template>
<script setup lang="ts">
import { RouterView } from "vue-router";
import LoaderWrapper from "./LoaderWrapper.vue";
import SidebarVue from "./sidebar/Sidebar.vue";
import HeaderVue from "./header/Header.vue";
import FooterPanel from "./footer/FooterPanel.vue";
import { NMessageProvider, useMessage } from "naive-ui";
import { defineComponent } from "vue";

// content
defineComponent({
  setup() {
    const message = useMessage();
    return {
      warning() {
        message.warning("...");
      },
    };
  },
});
</script>
