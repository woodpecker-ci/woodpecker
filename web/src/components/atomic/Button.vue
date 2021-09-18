<template>
  <button
    type="button"
    class="
      bg-white
      text-gray-800
      py-1
      px-4
      border border-gray-200
      rounded
      shadow-sm
      cursor-pointer
      focus:outline-transparent
      hover:bg-light-300
      disabled:opacity-50 disabled:cursor-not-allowed
    "
    :disabled="disabled"
    @click="doClick"
  >
    <slot>
      {{ text }}
    </slot>
  </button>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { RouteLocationRaw, useRouter } from 'vue-router';

export default defineComponent({
  name: 'Button',

  props: {
    text: {
      type: String,
      default: null,
    },

    disabled: {
      type: Boolean,
      required: false,
    },

    to: {
      type: [String, Object, null] as PropType<RouteLocationRaw | null>,
      default: null,
    },
  },

  setup(props) {
    const router = useRouter();

    async function doClick() {
      if (!props.to) {
        return;
      }

      if (typeof props.to === 'string' && props.to.startsWith('http')) {
        window.location.href = props.to;
        return;
      }

      await router.push(props.to);
    }

    return { doClick };
  },
});
</script>
