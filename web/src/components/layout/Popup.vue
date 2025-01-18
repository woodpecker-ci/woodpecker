<template>
  <!-- overlay -->
  <div
    v-if="open"
    class="fixed bottom-0 left-0 right-0 top-0 z-40 bg-gray-900 opacity-80 print:hidden"
    @click="$emit('close')"
  />
  <!-- overlay end -->
  <div v-if="open" class="fixed inset-0 z-50 m-auto flex max-w-2xl print:hidden">
    <div class="shadow-all m-auto flex h-auto flex-col p-2">
      <slot />
    </div>
  </div>
</template>

<script lang="ts" setup>
import { onKeyStroke } from '@vueuse/core';
import { toRef } from 'vue';

const props = defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const open = toRef(props, 'open');

onKeyStroke('Escape', (e) => {
  e.preventDefault();
  if (open.value) {
    emit('close');
  }
});
</script>
