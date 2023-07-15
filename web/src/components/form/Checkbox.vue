<template>
  <div class="flex items-center mb-2">
    <input
      :id="`checkbox-${id}`"
      type="checkbox"
      class="checkbox flex-shrink-0 relative border border-gray-400 cursor-pointer rounded-md transition-colors duration-150 w-5 h-5 checked:bg-woodpecker-400 checked:border-woodpecker-400 focus-visible:border-gray-600 dark:(border-gray-600 checked:bg-woodpecker-500 checked:border-woodpecker-500 focus-visible:border-gray-300 checked:focus-visible:border-gray-300)"
      :checked="innerValue"
      @click="innerValue = !innerValue"
    />
    <div class="flex flex-col ml-4">
      <label v-if="label" class="cursor-pointer text-color" :for="`checkbox-${id}`">{{ label }}</label>
      <span v-if="description" class="text-sm text-color-alt">{{ description }}</span>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, toRef } from 'vue';

export default defineComponent({
  name: 'Checkbox',

  props: {
    modelValue: {
      type: Boolean,
      required: true,
    },

    label: {
      type: String,
      default: null,
    },

    description: {
      type: String,
      default: null,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:modelValue': (_value: boolean): boolean => true,
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
  @apply dark:border-gray-300;
}

.checkbox:checked::before {
  opacity: 1;
}
</style>
