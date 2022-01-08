<template>
  <div class="w-full rounded-md overflow-hidden">
    <div class="font-bold bg-gray-400 dark:bg-dark-gray-800 p-2">{{ name }}</div>
    <div class="w-full p-2 font-mono whitespace-pre bg-gray-300 dark:bg-dark-gray-700">
      {{ text }}
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'CodeViewer',

  props: {
    name: {
      type: String,
      required: true,
    },

    text: {
      type: String,
      required: true,
    },
  },
});
</script>
