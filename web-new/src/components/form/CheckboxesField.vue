<template>
  <div v-for="option in options" :key="option.value" class="flex items-center mb-2">
    <input
      type="checkbox"
      class="
        checkbox
        relative
        border border-gray-400
        cursor-pointer
        rounded-md
        w-5
        h-5
        checked:bg-green checked:border-green checked:text-white
      "
      @click="clickOption(option)"
      :id="`checkbox-${id}-${option.value}`"
      :value="option.value"
      :checked="innerValue.includes(option.value)"
    />
    <label class="ml-4 cursor-pointer" :for="`checkbox-${id}-${option.value}`">{{ option.text }}</label>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, PropType } from 'vue';

import { CheckboxOption } from './form.types';

export default defineComponent({
  name: 'CheckboxField',

  components: {},

  props: {
    modelValue: {
      type: Array as PropType<CheckboxOption['value'][]>,
      default: [],
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
    const innerValue = computed({
      get: () => props.modelValue,
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

    const id = (Math.random() + 1).toString(36).substring(7);

    return {
      id,
      innerValue,
      clickOption,
    };
  },
});
</script>

<style scoped>
.checkbox {
  appearance: none;
  outline: 0;
  cursor: pointer;
  transition: background 175ms cubic-bezier(0.1, 0.1, 0.25, 1);
}

.checkbox::before {
  position: absolute;
  content: '';
  display: block;
  top: 50%;
  left: 50%;
  width: 8px;
  height: 14px;
  border-style: solid;
  border-color: white;
  border-width: 0 2px 2px 0;
  transform: translate(-50%, -60%) rotate(45deg);
  opacity: 0;
}

.checkbox:checked::before {
  opacity: 1;
}
</style>
