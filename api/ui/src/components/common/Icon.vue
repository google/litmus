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
  <component :is="componentName" v-bind="options">
    <slot v-if="$slots.default" />
    <template v-else>
      <Icon v-if="icon" :icon="icon" :width="size" :height="size" />
    </template>
  </component>
</template>

<script setup lang="ts">
 import { NIconWrapper, NIcon } from "naive-ui";
 import { Icon, loadIcon, type IconifyIcon } from "@iconify/vue";
 import { computed, ref, watchEffect } from "vue";

 const props = defineProps<{
name?: string;
size?: number;
bgSize?: number;
color?: string;
bgColor?: string;
borderRadius?: number;
depth?: 1 | 2 | 3 | 4 | 5;
 }>();

 const useWrapper = computed(() => !!(props.bgColor || props.bgSize || props.borderRadius));
 const componentName = computed(() => (useWrapper.value ? NIconWrapper : NIcon));

 const options = computed(() => {
const opt: Record<string, any> = {};
if (useWrapper.value) {
  if (props.bgSize !== undefined) opt.size = props.bgSize;
  if (props.bgColor !== undefined) opt.color = props.bgColor;
  if (props.borderRadius !== undefined) opt.borderRadius = props.borderRadius;
  if (props.color !== undefined) opt.iconColor = props.color;
} else {
  if (props.color !== undefined) opt.color = props.color;
  if (props.depth !== undefined) opt.depth = props.depth;
  if (props.size !== undefined) opt.size = props.size;
}
return opt;
 });

 const icon = ref<void | Required<IconifyIcon>>();

 const loadIconAsync = async (name: string) => {
try {
  const iconData = await loadIcon(name);
  icon.value = iconData;
} catch (error) {
  console.error(`Failed to load icon ${name}`, error);
}
 };

 watchEffect(() => {
if (props.name) {
  loadIconAsync(props.name);
} else {
  icon.value = undefined;
}
 });
</script>
