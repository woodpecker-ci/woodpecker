<template>
  <button
    type="button"
    class="
      py-1
      px-4
      rounded
      border
      shadow-sm
      cursor-pointer
      font-bold
      transition-colors
      duration-150
      focus:outline-none
      disabled:opacity-50 disabled:cursor-not-allowed
    "
    :class="{
      'bg-white hover:bg-gray-200 border-gray-300 text-gray-700': color === 'gray',
      'bg-lime-600 hover:bg-lime-700 border-lime-800 text-white': color === 'green',
      'bg-cyan-600 hover:bg-cyan-700 border-cyan-800 text-white': color === 'blue',
      'bg-red-500 hover:bg-red-600 border-red-700 text-white': color === 'red',
    }"
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

    color: {
      type: String as PropType<'blue' | 'green' | 'red' | 'gray'>,
      default: 'gray',
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
