<template>
  <svg fill="currentColor" :width="size" :height="size" viewBox="0 0 24 24">
    <circle v-if="bgCircle" cx="12" cy="12" :r="9" class="fill-transparent dark:fill-gray-300" />
    <path :d="path" />
  </svg>
</template>

<script lang="ts" setup>
defineProps<{
  path: string;
  size: number | string;
  bgCircle: boolean;
}>();
</script>
