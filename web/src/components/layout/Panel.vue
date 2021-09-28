<template>
  <div class="rounded-md w-full p-4 shadow border bg-white dark:bg-dark-gray-700 dark:border-dark-200">
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
