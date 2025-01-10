<template>
  <span class="inline-flex text-xs font-medium">
    <span
      class="flex items-center rounded-l-full border-2 border-wp-state-neutral-100 bg-wp-state-neutral-100 py-0.5 pl-2 pr-1 text-gray-300"
      :class="{
        'rounded-r-full pr-2': value === undefined,
      }"
    >
      {{ label }}
    </span>
    <span
      v-if="value !== undefined"
      class="flex items-center rounded-r-full border-2 border-wp-state-neutral-100 py-0.5 pl-1 pr-2"
    >
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
