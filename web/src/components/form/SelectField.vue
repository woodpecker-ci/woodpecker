<template>
  <select
    v-model="innerValue"
    class="border-wp-control-neutral-200 bg-wp-control-neutral-100 text-wp-text-100 w-full rounded-md border px-2 py-1"
  >
    <option v-if="placeholder" value="" class="hidden">{{ placeholder }}</option>
    <option v-for="option in options" :key="option.value" :value="option.value" class="text-wp-text-100">
      {{ option.text }}
    </option>
  </select>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import type { SelectOption } from './form.types';

const props = defineProps<{
  modelValue: string;
  placeholder?: string;
  options: SelectOption[];
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void;
}>();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value,
  set: (selectedValue) => {
    emit('update:modelValue', selectedValue);
  },
});
</script>
