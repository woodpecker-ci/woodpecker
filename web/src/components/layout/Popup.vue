<template>
  <!-- overlay -->
  <div
    v-if="open"
    class="fixed bg-gray-900 opacity-80 left-0 top-0 right-0 bottom-0 z-500 print:hidden"
    @click="$emit('close')"
  />
  <!-- overlay end -->
  <div
    v-if="open"
    class="print:hidden fixed flex max-w-1/3 <md:max-w-4/5 max-h-3/5 top-1/2 left-1/2 transform z-1000 -translate-x-1/2 -translate-y-1/2"
  >
    <div class="m-auto flex flex-col shadow-all h-auto">
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
