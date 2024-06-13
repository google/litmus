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
  <v-navigation-drawer
    left
    v-model="customizer.sidebar_drawer"
    elevation="0"
    rail-width="60"
    mobile-breakpoint="lg"
    app
    class="leftSidebar"
    :rail="customizer.mini_sidebar"
    expand-on-hover
  >
    <div class="pa-5">
      <Logo />
    </div>
    <PerfectScrollbar class="scrollnavbar">
      <v-list aria-busy="true" aria-label="menu list">
        <template v-for="(item, i) in sidebarMenu" :key="i">
          <NavigationGroup v-if="item.header" :item="item" :key="i" />
          <v-divider v-else-if="item.divider" class="my-3" />
          <NavigationCollapse
            v-else-if="item.children"
            :item="item"
            :level="0"
            class="leftPadding"
          />
          <NavigationItem v-else :item="item" />
        </template>
      </v-list>
    </PerfectScrollbar>
  </v-navigation-drawer>
</template>

<script setup lang="ts">
import { shallowRef } from "vue";
import { useCustomizerStore } from "../../../stores/customizer";
import sidebarItems from "./sidebarItem";

import NavigationGroup from "./NavigationGroup.vue";
import NavigationItem from "./NavigationItem.vue";
import NavigationCollapse from "./NavigationCollapse.vue";
import Logo from "../logo/logo.vue";

const customizer = useCustomizerStore();
const sidebarMenu = shallowRef(sidebarItems);
</script>
