<template>
  <div
    class="
      text-sm text-gray-600
      font-bold
      rounded-md
      border border-solid
      p-2
      border-yellow-500
      bg-yellow-200
      dark:bg-yellow-600 dark:border-yellow-800 dark:text-light-100
    "
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
