<template>
  <div class="rounded-md w-full shadow overflow-hidden text-gray-500 bg-gray-300 dark:bg-dark-gray-700">
    <div v-if="title" class="font-bold bg-gray-400 dark:bg-dark-gray-800 p-2">{{ title }}</div>
    <div class="w-full p-4 bg-gray-300 dark:bg-dark-gray-700">
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
