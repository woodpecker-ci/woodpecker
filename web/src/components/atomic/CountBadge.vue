<template>
  <span
    class="inline-block bg-wp-background-400 dark:bg-wp-background-200 px-1.5 py-0.5 rounded-full min-w-5 font-bold text-center text-wp-text-100 text-xs leading-4"
  >
    {{ value }}
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  value?: string | number;
}>();
</script>
