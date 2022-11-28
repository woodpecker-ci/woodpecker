<template>
  <input
    v-if="lines === 1"
    v-model="innerValue"
    class="w-full border border-gray-200 py-1 px-2 rounded-md bg-white dark:bg-dark-gray-700 dark:border-dark-400 focus-visible:outline-none focus-visible:border-gray-600 dark:focus-visible:border-gray-300"
    :class="{ 'bg-gray-200 dark:bg-gray-600': disabled }"
    :disabled="disabled"
    :type="type"
    :placeholder="placeholder"
  />
  <textarea
    v-else
    v-model="innerValue"
    class="w-full border border-gray-200 py-1 px-2 rounded-md bg-white dark:bg-dark-gray-700 dark:border-dark-400 focus-visible:outline-none focus-visible:border-gray-600 dark:focus-visible:border-gray-300"
    :class="{ 'bg-gray-200 dark:bg-gray-600': disabled }"
    :disabled="disabled"
    :placeholder="placeholder"
    :rows="lines"
  />
</template>

<script lang="ts">
import { computed, defineComponent, toRef } from 'vue';

export default defineComponent({
  name: 'TextField',

  props: {
    modelValue: {
      type: String,
      default: '',
    },

    placeholder: {
      type: String,
      default: '',
    },

    type: {
      type: String,
      default: 'text',
    },

    lines: {
      type: Number,
      default: 1,
    },

    disabled: {
      type: Boolean,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: string): boolean => true,
  },

  setup: (props, ctx) => {
    const modelValue = toRef(props, 'modelValue');
    const innerValue = computed({
      get: () => modelValue.value,
      set: (value) => {
        ctx.emit('update:modelValue', value);
      },
    });

    return {
      innerValue,
    };
  },
});
</script>
