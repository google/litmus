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
  <v-list-group no-action>
    <template v-slot:activator="{ props }">
      <v-list-item
        v-bind="props"
        :value="item.title"
        rounded
        class="mb-1"
        color="primary"
      >
        <template v-slot:prepend>
          <component :is="item.icon" class="iconClass" :level="level"></component>
        </template>
        <v-list-item-title class="mr-auto">{{ item.title }}</v-list-item-title>
        <v-list-item-subtitle v-if="item.subCaption" class="text-caption mt-n1 hide-menu">
          {{ item.subCaption }}
        </v-list-item-subtitle>
      </v-list-item>
    </template>
    <template v-for="(subitem, i) in item.children" :key="i">
      <NavigationCollapse v-if="subitem.children" :item="subitem" :level="level + 1" />
      <NavigationItem v-else :item="subitem" :level="level + 1" />
    </template>
  </v-list-group>
</template>

<script setup>
import NavigationItem from "./NavigationItem.vue";
import NavigationCollapse from "./NavigationCollapse.vue";

const props = defineProps({ item: Object, level: Number });
</script>
