<template>
  <div
    class="text-sm text-gray-600 font-bold rounded-md border border-solid p-2 border-wp-control-error-300 bg-wp-control-error-100"
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
