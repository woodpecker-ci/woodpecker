<template>
  <span class="inline-flex items-center text-xs font-medium">
    <span
      v-if="label !== undefined"
      class="border-wp-state-neutral-100 bg-wp-state-neutral-100 rounded-l-full border-1 py-0.5 pr-1 pl-2 whitespace-nowrap text-gray-300"
      :class="{
        'rounded-r-full pr-2': value === undefined,
      }"
    >
      {{ label }}
    </span>
    <span
      v-if="value !== undefined"
      class="border-wp-state-neutral-100 rounded-r-full border-1 py-0.5 pr-2 pl-1 whitespace-nowrap"
      :class="{
        'rounded-l-full pl-2': label === undefined,
      }"
    >
      {{ value }}
    </span>
  </span>
</template>

<script lang="ts" setup>
defineProps<{
  label?: string;
  value?: string | number;
}>();
</script>
