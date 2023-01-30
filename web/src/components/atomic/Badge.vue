<template>
  <span class="text-xs font-medium inline-flex">
    <span
      class="pl-2 pr-1 py-0.5 bg-gray-800 text-gray-200 dark:bg-gray-600 border-2 border-gray-800 dark:border-gray-600 rounded-l-full"
    >
      {{ label }}
    </span>
    <span class="pl-1 pr-2 py-0.5 border-2 border-gray-800 dark:border-gray-600 rounded-r-full">
      {{ value }}
    </span>
  </span>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';

const props = defineProps<{
  label: string;
  value: string | number;
}>();

const label = toRef(props, 'label');
const value = toRef(props, 'value');
</script>
