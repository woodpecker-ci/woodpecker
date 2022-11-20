<template>
  <div class="rounded-md w-full shadow overflow-hidden bg-gray-300 dark:bg-dark-gray-700">
    <div v-if="title" class="font-bold text-gray-200 bg-gray-400 dark:bg-dark-gray-800 px-4 py-2">{{ title }}</div>
    <div class="w-full p-4 bg-white dark:bg-dark-gray-700 text-color">
      <slot />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Panel',

  props: {
    title: {
      type: String,
      default: '',
    },
  },
});
</script>
