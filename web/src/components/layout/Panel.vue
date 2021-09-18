<template>
  <div class="rounded-md w-full p-4 shadow border bg-white">
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Panel',

  setup() {
    return {};
  },
});
</script>
