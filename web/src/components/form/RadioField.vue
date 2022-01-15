<template>
  <div v-for="option in options" :key="option.value" class="flex items-center mb-2">
    <input
      :id="`radio-${id}-${option.value}`"
      type="radio"
      class="
        radio
        relative
        flex-shrink-0
        border border-gray-400
        dark:border-gray-600
        cursor-pointer
        rounded-full
        w-5
        h-5
        checked:bg-lime-600 checked:border-lime-600
        dark:checked:bg-lime-700 dark:checked:border-lime-700
      "
      :value="option.value"
      :checked="innerValue.includes(option.value)"
      @click="innerValue = option.value"
    />
    <div class="flex flex-col ml-4">
      <label class="cursor-pointer text-gray-600 dark:text-gray-500" :for="`radio-${id}-${option.value}`">{{
        option.text
      }}</label>
      <span v-if="option.description" class="text-sm text-gray-400 dark:text-gray-600">{{ option.description }}</span>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, toRef } from 'vue';

import { RadioOption } from './form.types';

export default defineComponent({
  name: 'RadioField',

  components: {},

  props: {
    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    modelValue: {
      type: String,
      required: true,
    },

    options: {
      type: Array as PropType<RadioOption[]>,
      required: true,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: RadioOption['value']): boolean => true,
  },

  setup: (props, ctx) => {
    const modelValue = toRef(props, 'modelValue');
    const innerValue = computed({
      get: () => modelValue.value,
      set: (value) => {
        ctx.emit('update:modelValue', value);
      },
    });

    const id = (Math.random() + 1).toString(36).substring(7);

    return {
      id,
      innerValue,
    };
  },
});
</script>

<style scoped>
.radio {
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
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: white;
  transform: translate(-50%, -50%);
  opacity: 0;
  @apply dark:bg-gray-400;
}

.radio:checked::before {
  opacity: 1;
}
</style>
