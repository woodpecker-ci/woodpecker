<template>
  <component
    :is="to ? 'router-link' : clickable ? 'button' : 'div'"
    :to="to"
    class="flex w-full overflow-hidden rounded-md border border-wp-background-400 bg-wp-background-100 p-4 dark:bg-wp-background-200"
    :class="{
      'cursor-pointer hover:bg-wp-background-300 hover:shadow-md dark:hover:bg-wp-background-300': clickable || to,
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
