<template>
  <component
    :is="to ? 'router-link' : clickable ? 'button' : 'div'"
    :to="to"
    class="flex border-wp-background-400 bg-wp-background-100 dark:bg-wp-background-200 border rounded-md w-full overflow-hidden"
    :class="{
      'cursor-pointer hover:shadow-md hover:bg-wp-background-300 dark:hover:bg-wp-background-300': clickable || to,
    }"
  >
    <slot />
  </component>
</template>

<script lang="ts" setup>
import type { RouteLocationRaw } from 'vue-router';

defineProps<{
  clickable?: boolean;
  to?: RouteLocationRaw;
}>();
</script>
