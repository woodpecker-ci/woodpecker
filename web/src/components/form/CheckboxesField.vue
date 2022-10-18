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

<script lang="ts">
import { computed, defineComponent, PropType, toRef } from 'vue';

import Checkbox from './Checkbox.vue';
import { CheckboxOption } from './form.types';

export default defineComponent({
  name: 'CheckboxesField',

  components: { Checkbox },

  props: {
    modelValue: {
      type: Array as PropType<CheckboxOption['value'][]>,
      default: () => [],
    },

    options: {
      type: Array as PropType<CheckboxOption[]>,
      required: true,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: CheckboxOption['value'][]): boolean => true,
  },

  setup: (props, ctx) => {
    const modelValue = toRef(props, 'modelValue');
    const innerValue = computed({
      get: () => modelValue.value,
      set: (value) => {
        ctx.emit('update:modelValue', value);
      },
    });

    function clickOption(option: CheckboxOption) {
      if (innerValue.value.includes(option.value)) {
        innerValue.value = innerValue.value.filter((o) => o !== option.value);
      } else {
        innerValue.value.push(option.value);
      }
    }

    return {
      innerValue,
      clickOption,
    };
  },
});
</script>
