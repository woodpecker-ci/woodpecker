<template>
  <component
    :is="to ? 'router-link' : clickable ? 'control' : 'div'"
    :to="to"
    class="w-full flex border rounded-md bg-wp-background-100 overflow-hidden p-4 border-wp-background-400 dark:bg-wp-background-200"
    :class="{
      'cursor-pointer hover:shadow-md hover:bg-wp-background-300 dark:hover:bg-wp-background-300': clickable || to,
    }"
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
