<template>
  <div
    class="text-sm text-wp-gray-600 font-bold rounded-md border border-solid p-2 border-wp-yellow-300 bg-wp-yellow-100 dark:bg-wp-yellow-600 dark:border-wp-yellow-800 dark:text-light-100"
  >
    ⚠ {{ text }}
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'Warning',

  props: {
    text: {
      type: String,
      required: true,
    },
  },
});
</script>
