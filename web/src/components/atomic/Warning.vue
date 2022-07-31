<template>
  <div class="text-sm font-bold text-light-600 rounded-md border border-solid p-2 border-red-900 bg-red-600">
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
