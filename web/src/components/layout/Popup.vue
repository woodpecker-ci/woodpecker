<template>
  <!-- overlay -->
  <div
    v-if="open"
    class="fixed bg-gray-900 opacity-80 left-0 top-0 right-0 bottom-0 z-500 print:hidden"
    @click="$emit('close')"
  />
  <!-- overlay end -->
  <transition class="print:hidden fixed flex left-0 top-0 right-0 bottom-0">
    <div v-if="open" class="m-auto flex flex-col shadow-all z-1000 max-w-3/5 max-h-3/5">
      <div class="m-auto flex">
        <slot />
      </div>
    </div>
  </transition>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Popup',

  props: {
    open: {
      type: Boolean,
    },
  },

  emits: {
    close: () => true,
  },
});
</script>
