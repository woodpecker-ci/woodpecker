<template>
  <select
    v-model="innerValue"
    class="bg-wp-control-neutral-100 text-wp-text-100 border-wp-control-neutral-200 w-full border py-1 px-2 rounded-md"
  >
    <option v-if="placeholder" value="" class="hidden">{{ placeholder }}</option>
    <option v-for="option in options" :key="option.value" :value="option.value" class="text-wp-text-100">
      {{ option.text }}
    </option>
  </select>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import { SelectOption } from './form.types';

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
