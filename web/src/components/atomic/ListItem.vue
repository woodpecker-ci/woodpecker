<template>
  <component
    :is="to ? 'router-link' : clickable ? 'button' : 'div'"
    :to="to"
    class="w-full flex border rounded-md bg-white overflow-hidden p-4 border-gray-300 dark:bg-dark-gray-700 dark:border-dark-400"
    :class="{ 'cursor-pointer hover:shadow-md hover:bg-gray-200 dark:hover:bg-dark-gray-800': clickable || to }"
  >
    <slot />
  </component>
</template>

<script lang="ts" setup>
import { RouteLocationRaw } from 'vue-router';

defineProps<{
  clickable?: boolean;
  to?: RouteLocationRaw;
}>();
</script>
