<template>
  <div
    class="text-sm text-gray-700 font-bold rounded-md border border-solid p-2 border-wp-hint-warn-200 bg-wp-hint-warn-100"
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
