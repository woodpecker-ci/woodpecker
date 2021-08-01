<template>
  <Button
    :disabled="disabled"
    :to="to"
    class="flex text-gray-400 px-0 py-0 hover:bg-transparent hover:text-gray-500 shadow-none bg-transparent border-none"
  >
    <slot />
  </Button>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { RouteLocationRaw, useRouter } from 'vue-router';
import Button from './Button.vue';

export default defineComponent({
  name: 'IconButton',

  components: { Button },

  props: {
    disabled: {
      type: Boolean,
      required: false,
    },

    to: {
      type: [String, Object] as PropType<RouteLocationRaw>,
      required: false,
    },
  },
});
</script>
