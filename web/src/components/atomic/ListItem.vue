<template>
  <component
    :is="to ? 'router-link' : clickable ? 'button' : 'div'"
    :to="to"
    class="border-wp-background-400 bg-wp-background-100 dark:bg-wp-background-200 flex w-full overflow-hidden rounded-md border p-4"
    :class="{
      'hover:bg-wp-background-300 dark:hover:bg-wp-background-300 cursor-pointer hover:shadow-md': clickable || to,
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
