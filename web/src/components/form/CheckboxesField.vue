<template>
  <Checkbox
    v-for="option in options"
    :key="option.value"
    :model-value="innerValue.includes(option.value)"
    :label="option.text"
    :disabled="disabled ? disabled(option) : false"
    :description="option.description"
    class="mb-2"
    @update:model-value="clickOption(option)"
  />
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Checkbox from './Checkbox.vue';
import type { CheckboxOption } from './form.types';

const props = defineProps<{
  modelValue?: CheckboxOption['value'][];
  options?: CheckboxOption[];
  disabled?: (option: CheckboxOption) => boolean;
}>();
const emit = defineEmits<{
  (event: 'update:modelValue', value: CheckboxOption['value'][]): void;
}>();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value || [],
  set: (value) => {
    emit('update:modelValue', value);
  },
});

function clickOption(option: CheckboxOption) {
  // Both branches must assign rather than mutate. Pushing onto the array the
  // getter returned skips the setter, so no update:modelValue is emitted, and a
  // parent whose v-model is a computed never sees the value.
  if (innerValue.value.includes(option.value)) {
    innerValue.value = innerValue.value.filter((o) => o !== option.value);
  } else {
    innerValue.value = [...innerValue.value, option.value];
  }
}
</script>
