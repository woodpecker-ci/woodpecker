<template>
  <Button
    :disabled="disabled"
    :to="to"
    class="
      flex
      items-center
      justify-center
      text-gray-400
      px-0
      py-0
      hover:bg-transparent hover:text-gray-500
      shadow-none
      bg-transparent
      border-none
    "
  >
    <Icon :name="icon" />
  </Button>
</template>

<script lang="ts">
import { defineComponent, ExtractDefaultPropTypes, PropType } from 'vue';
import { RouteLocationRaw } from 'vue-router';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';

export default defineComponent({
  name: 'IconButton',

  components: { Button, Icon },

  props: {
    icon: {
      type: String as ExtractDefaultPropTypes<typeof Icon>['name'], // TODO
      required: true,
    },

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
