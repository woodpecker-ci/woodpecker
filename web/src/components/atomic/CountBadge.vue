<template>
  <span
    class="bg-wp-code-inline-100 text-wp-text-code-inline-100 inline-block px-1.5 py-0.5 rounded-full min-w-5 font-bold text-center text-xs leading-4"
  >
    {{ value }}
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  value?: string | number;
}>();
</script>
