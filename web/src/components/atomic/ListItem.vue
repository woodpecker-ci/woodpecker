<template>
  <component
    :is="to ? 'router-link' : clickable ? 'button' : 'div'"
    :to="to"
    class="w-full flex border rounded-md bg-white overflow-hidden p-4 border-wp-gray-300 dark:bg-wp-darkgray-700 dark:border-wp-dark-400"
    :class="{ 'cursor-pointer hover:shadow-md hover:bg-wp-gray-200 dark:hover:bg-wp-darkgray-800': clickable || to }"
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
