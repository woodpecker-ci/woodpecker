<template>
  <span class="text-xs font-medium inline-flex">
    <span
      class="pl-2 pr-1 py-0.5 bg-wp-state-neutral-100 text-gray-300 border-2 border-wp-state-neutral-100 rounded-l-full"
      :class="{
        'rounded-r-full pr-2': value === undefined,
      }"
    >
      {{ label }}
    </span>
    <span v-if="value !== undefined" class="pl-1 pr-2 py-0.5 border-2 border-wp-state-neutral-100 rounded-r-full">
      {{ value }}
    </span>
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  label: string;
  value?: string | number;
}>();
</script>
