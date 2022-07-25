<template>
  <component
    :is="clickable ? 'button' : 'div'"
    class="
      w-full
      flex
      border
      rounded-md
      bg-white
      overflow-hidden
      p-4
      border-gray-300
      dark:bg-dark-gray-700 dark:border-dark-400
    "
    :class="{ 'cursor-pointer hover:shadow-md hover:bg-gray-200 dark:hover:bg-dark-gray-800': clickable }"
  >
    <slot />
  </component>
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
