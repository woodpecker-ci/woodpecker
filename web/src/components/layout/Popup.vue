<template>
  <!-- overlay -->
  <div
    v-if="open"
    class="fixed bottom-0 left-0 right-0 top-0 z-40 bg-gray-900 opacity-80 print:hidden"
    @click="$emit('close')"
  />
  <!-- overlay end -->
  <div
    v-if="open"
    class="max-w-1/3 max-w-4/5 md:max-h-3/5 fixed left-1/2 top-1/2 z-50 flex -translate-x-1/2 -translate-y-1/2 transform print:hidden"
  >
    <div class="shadow-all m-auto flex h-auto flex-col">
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
