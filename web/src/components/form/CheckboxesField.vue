<template>
  <Checkbox
    v-for="option in options"
    :key="option.value"
    :model-value="innerValue.includes(option.value)"
    :label="option.text"
    :description="option.description"
    class="mb-2"
    @update:model-value="clickOption(option)"
  />
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Checkbox from './Checkbox.vue';
import { CheckboxOption } from './form.types';

const props = withDefaults(
  defineProps<{
    modelValue: CheckboxOption['value'][];
    options: CheckboxOption[];
  }>(),
  {
    modelValue: () => [],
    options: undefined,
  },
);

const emit = defineEmits<{
  (event: 'update:modelValue', value: CheckboxOption['value'][]): void;
}>();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value,
  set: (value) => {
    emit('update:modelValue', value);
  },
});

function clickOption(option: CheckboxOption) {
  if (innerValue.value.includes(option.value)) {
    innerValue.value = innerValue.value.filter((o) => o !== option.value);
  } else {
    innerValue.value.push(option.value);
  }
}
</script>
