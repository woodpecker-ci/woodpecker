<template>
  <div v-for="option in options" :key="option.value" class="mb-2 flex items-center">
    <input
      :id="`radio-${id}-${option.value}`"
      type="radio"
      class="radio relative h-5 w-5 flex-shrink-0 cursor-pointer rounded-full border border-wp-control-neutral-200 bg-wp-control-neutral-100 checked:border-wp-control-ok-200 checked:bg-wp-control-ok-200 focus-visible:border-wp-control-neutral-300 checked:focus-visible:border-wp-control-ok-300"
      :value="option.value"
      :checked="innerValue?.includes(option.value)"
      @click="innerValue = option.value"
    />
    <div class="ml-4 flex flex-col">
      <label class="text-wp-text-100 cursor-pointer" :for="`radio-${id}-${option.value}`">{{ option.text }}</label>
      <span v-if="option.description" class="text-wp-text-alt-100 text-sm">{{ option.description }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import type { RadioOption } from './form.types';

const props = defineProps<{
  modelValue: string;
  options: RadioOption[];
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: string): void;
}>();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value,
  set: (value) => {
    emit('update:modelValue', value);
  },
});

const id = (Math.random() + 1).toString(36).substring(7);
</script>

<style scoped>
.radio {
  width: 1.3rem;
  height: 1.3rem;
  appearance: none;
  outline: 0;
  cursor: pointer;
  transition: background 175ms cubic-bezier(0.1, 0.1, 0.25, 1);
}

.radio::before {
  position: absolute;
  content: '';
  display: block;
  top: 50%;
  left: 50%;
  width: 0.5rem;
  height: 0.5rem;
  border-radius: 50%;
  background: white;
  transform: translate(-50%, -50%);
  opacity: 0;
  @apply dark:bg-white;
}

.radio:checked::before {
  opacity: 1;
}
</style>
