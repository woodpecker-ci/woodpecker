<template>
  <input
    v-if="lines === 1"
    v-model="innerValue"
    class="w-full border border-wp-control-neutral-200 py-1 px-2 rounded-md bg-wp-background-100 focus-visible:outline-none focus-visible:border-wp-control-neutral-300"
    :class="{ 'opacity-50': disabled }"
    :disabled="disabled"
    :type="type"
    :placeholder="placeholder"
  />
  <textarea
    v-else
    v-model="innerValue"
    class="w-full border border-wp-control-neutral-200 py-1 px-2 rounded-md bg-wp-background-100 focus-visible:outline-none focus-visible:border-wp-control-neutral-300"
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
    modelValue: string;
    placeholder: string;
    type: string;
    lines: number;
    disabled?: boolean;
  }>(),
  {
    modelValue: '',
    placeholder: '',
    type: 'text',
    lines: 1,
  },
);

const emit = defineEmits(['update:modelValue']);

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value,
  set: (value) => {
    emit('update:modelValue', value);
  },
});
</script>
