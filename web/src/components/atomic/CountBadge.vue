<template>
  <span
    class="inline-block min-w-5 rounded-full bg-wp-background-300 px-1.5 py-0.5 text-center text-xs font-bold leading-4 text-wp-text-100 dark:bg-wp-background-100"
  >
    {{ value }}
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  value?: string | number;
}>();
</script>
