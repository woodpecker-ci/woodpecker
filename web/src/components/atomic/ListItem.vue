<template>
  <component
    :is="to ? 'router-link' : clickable ? 'button' : 'div'"
    :to="to"
    class="border-wp-background-300 dark:border-wp-background-100 bg-wp-background-200 dark:bg-wp-background-200 flex w-full overflow-hidden rounded-md border p-4"
    :class="{
      'hover:bg-wp-control-neutral-100 dark:hover:bg-wp-control-neutral-200 cursor-pointer': clickable || to,
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
