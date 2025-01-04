<template>
  <input
    v-if="lines === 1"
    v-model="innerValue"
    class="w-fill rounded-md border border-wp-control-neutral-200 bg-wp-background-100 px-2 py-1 focus-visible:border-wp-control-neutral-300 focus-visible:outline-none"
    :class="{ 'opacity-50': disabled }"
    :disabled="disabled"
    :type="type"
    :placeholder="placeholder"
  />
  <textarea
    v-else
    v-model="innerValue"
    class="w-fill rounded-md border border-wp-control-neutral-200 bg-wp-background-100 px-2 py-1 focus-visible:border-wp-control-neutral-300 focus-visible:outline-none"
    :class="{ 'opacity-50': disabled }"
    :disabled="disabled"
    :placeholder="placeholder"
    :rows="lines"
  />
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

const props = withDefaults(
  defineProps<{
    modelValue?: string;
    placeholder?: string;
    type?: string;
    lines?: number;
    disabled?: boolean;
  }>(),
  {
    modelValue: '',
    placeholder: '',
    type: 'text',
    lines: 1,
  },
);

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
</script>
