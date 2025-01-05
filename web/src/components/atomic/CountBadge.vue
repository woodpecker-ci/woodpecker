<template>
  <span
    class="bg-wp-background-300 dark:bg-wp-background-100 text-wp-text-100 inline-block min-w-5 rounded-full px-1.5 py-0.5 text-center text-xs font-bold leading-4"
  >
    {{ value }}
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  value?: string | number;
}>();
</script>
