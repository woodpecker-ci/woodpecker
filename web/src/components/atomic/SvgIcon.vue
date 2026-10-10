<template>
  <svg fill="currentColor" :width="size" :height="size" viewBox="0 0 24 24">
    <path :d="path" />
  </svg>
</template>

<script lang="ts" setup>
defineProps<{
  path: string;
  size: number | string;
}>();
</script>
