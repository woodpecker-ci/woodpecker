<template>
  <TextField v-model="innerValue" :placeholder="placeholder" type="number" />
</template>

<script lang="ts" setup>
import TextField from '~/components/form/TextField.vue';
import { computed, toRef } from 'vue';

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
    emit('update:modelValue', Number.parseFloat(value));
  },
});
</script>
