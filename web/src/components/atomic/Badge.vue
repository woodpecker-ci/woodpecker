<template>
  <span class="text-xs font-medium inline-flex">
    <span class="bg-gray-500 dark:bg-dark-700 dark:text-gray-400 text-white pl-2 pr-1 py-0.5 rounded-l">
      {{ label }}
    </span>
    <span class="bg-blue-600 text-blue-100 dark:bg-blue-800 dark:text-blue-300 pl-1 pr-2 py-0.5 rounded-r">
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
