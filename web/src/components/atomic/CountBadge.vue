<template>
  <span
    class="text-xs font-bold inline-block min-w-5 leading-4 rounded-full bg-wp-background-300 dark:bg-wp-background-100 text-wp-text-100 text-center py-0.5 px-1.5"
  >
    {{ value }}
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  label: string;
  value?: string | number;
}>();
</script>
