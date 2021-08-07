<template>
  <div class="flex rounded-md w-full p-4 shadow border bg-white">
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent, provide, ref } from 'vue';

export default defineComponent({
  name: 'Panel',

  setup() {
    return {};
  },
});
</script>
