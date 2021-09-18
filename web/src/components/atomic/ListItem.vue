<template>
  <div
    class="w-full flex border rounded-md bg-white overflow-hidden p-4"
    :class="{ 'cursor-pointer hover:shadow-sm hover:bg-light-200': clickable }"
  >
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'ListItem',

  props: {
    clickable: {
      type: Boolean,
    },
  },
});
</script>
