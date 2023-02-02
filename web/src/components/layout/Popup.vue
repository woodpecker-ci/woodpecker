<template>
  <!-- overlay -->
  <div
    v-if="open"
    class="fixed bg-gray-900 opacity-80 left-0 top-0 right-0 bottom-0 z-500 print:hidden"
    @click="$emit('close')"
  />
  <!-- overlay end -->
  <transition class="print:hidden fixed flex top-1/2 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
    <div v-if="open" class="m-auto flex flex-col shadow-all z-1000 max-w-3/5 max-h-3/5 h-auto">
      <slot />
    </div>
  </transition>
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
