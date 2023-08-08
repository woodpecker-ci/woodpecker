<template>
  <TextField v-model="innerValue" :placeholder="placeholder" type="number" />
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import TextField from '~/components/form/TextField.vue';

const props = defineProps<{
  modelValue: number;
  placeholder?: string;
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: number): void;
}>();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value.toString(),
  set: (value) => {
    emit('update:modelValue', parseFloat(value));
  },
});
</script>
