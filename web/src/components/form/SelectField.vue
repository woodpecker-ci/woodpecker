<template>
  <select
    v-model="innerValue"
    class="bg-wp-control-neutral-100 text-wp-text-100 border-wp-control-neutral-200 w-full border py-1 px-2 rounded-md"
    :class="{
      'text-wp-text-100': innerValue === '',
      'text-wp-gray-900': innerValue !== '',
    }"
  >
    <option v-if="placeholder" value="" class="hidden">{{ placeholder }}</option>
    <option v-for="option in options" :key="option.value" :value="option.value" class="text-wp-text-100">
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
