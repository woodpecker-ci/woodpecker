<template>
  <div
    class="flex gap-2 items-center text-sm text-gray-700 font-bold rounded-md border border-solid p-2 border-wp-hint-warn-200 bg-wp-hint-warn-100"
  >
    <Icon name="warning" />
    <span>{{ text }}</span>
  </div>
</template>

<script lang="ts" setup>
defineProps<{
  text: string;
}>();
</script>
