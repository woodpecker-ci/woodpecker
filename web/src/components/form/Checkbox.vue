<template>
  <div class="flex items-center mb-2">
    <input
      :id="`checkbox-${id}`"
      type="checkbox"
      class="relative flex-shrink-0 border-wp-control-neutral-200 checked:border-wp-control-ok-200 focus-visible:border-wp-control-neutral-300 checked:focus-visible:border-wp-control-ok-300 bg-wp-control-neutral-100 checked:bg-wp-control-ok-200 border rounded-md w-5 h-5 transition-colors duration-150 cursor-pointer checkbox"
      :checked="innerValue"
      @click="innerValue = !innerValue"
    />
    <div class="flex flex-col ml-4">
      <label class="text-wp-text-100 cursor-pointer" :for="`checkbox-${id}`">{{ label }}</label>
      <span v-if="description" class="text-sm text-wp-text-alt-100">{{ description }}</span>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

const props = defineProps<{
  modelValue: boolean;
  label: string;
  description?: string;
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: boolean): void;
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
.checkbox {
  width: 1.5rem;
  height: 1.5rem;
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
  width: 0.5rem;
  height: 1rem;
  border-style: solid;
  border-color: white;
  border-width: 0 2px 2px 0;
  transform: translate(-50%, -60%) rotate(45deg);
  opacity: 0;
  @apply dark:border-white;
}

.checkbox:checked::before {
  opacity: 1;
}
</style>
