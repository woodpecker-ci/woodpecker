<template>
  <select
    v-model="innerValue"
    class="w-full border border-gray-900 py-1 px-2 rounded-md bg-white focus:outline-none border-gray-900"
    :class="{
      'text-color': innerValue === '',
      'text-gray-900': innerValue !== '',
    }"
  >
    <option v-if="placeholder" value="" class="hidden">{{ placeholder }}</option>
    <option v-for="option in options" :key="option.value" :value="option.value" class="text-color">
      {{ option.text }}
    </option>
  </select>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, toRef } from 'vue';

import { SelectOption } from './form.types';

export default defineComponent({
  name: 'SelectField',

  props: {
    modelValue: {
      type: String,
      default: null,
    },

    placeholder: {
      type: String,
      default: null,
    },

    options: {
      type: Array as PropType<SelectOption[]>,
      required: true,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: SelectOption['value'] | null): boolean => true,
  },

  setup: (props, ctx) => {
    const modelValue = toRef(props, 'modelValue');
    const innerValue = computed({
      get: () => modelValue.value,
      set: (selectedValue) => {
        ctx.emit('update:modelValue', selectedValue);
      },
    });

    return {
      innerValue,
    };
  },
});
</script>
