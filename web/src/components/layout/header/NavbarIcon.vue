<template>
  <button type="button" :title="title" :aria-label="title" class="navbar-icon" @click="doClick">
    <slot />
  </button>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { RouteLocationRaw, useRouter } from 'vue-router';

export default defineComponent({
  name: 'NavbarIcon',

  props: {
    to: {
      type: [String, Object, null] as PropType<RouteLocationRaw | null>,
      default: null,
    },

    title: {
      type: String,
      required: true,
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

<style scoped>
.navbar-icon {
  @apply w-11 h-11 rounded-full p-2.5;
}

.navbar-icon :deep(svg) {
  @apply w-full h-full;
}
</style>
