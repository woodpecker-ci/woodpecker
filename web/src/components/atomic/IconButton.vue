<template>
  <router-link v-if="to" :to="to" :title="title" :aria-label="title" class="icon-button">
    <slot>
      <Icon :name="icon" />
    </slot>
  </router-link>
  <button
    v-else
    :disabled="disabled"
    class="icon-button"
    type="button"
    :title="title"
    :aria-label="title"
    @click="doClick"
  >
    <slot>
      <Icon :name="icon" />
    </slot>
    <div v-if="isLoading" class="absolute left-0 top-0 right-0 bottom-0 flex items-center justify-center">
      <Icon name="loading" class="animate-spin" />
    </div>
  </button>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { RouteLocationRaw, useRouter } from 'vue-router';

import Icon, { IconNames } from '~/components/atomic/Icon.vue';

export default defineComponent({
  name: 'IconButton',

  components: { Icon },

  props: {
    icon: {
      type: String as PropType<IconNames>,
      default: '',
    },

    disabled: {
      type: Boolean,
      required: false,
    },

    to: {
      type: [String, Object, null] as PropType<RouteLocationRaw | null>,
      default: null,
    },

    isLoading: {
      type: Boolean,
    },

    title: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const router = useRouter();

    async function doClick() {
      if (props.isLoading) {
        return;
      }

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
.icon-button {
  @apply relative flex items-center justify-center px-1 py-1 rounded-full bg-transparent hover:bg-gray-200 dark:hover:bg-gray-600 cursor-pointer transition-all duration-150 focus:outline-none overflow-hidden disabled:opacity-50 disabled:cursor-not-allowed;
}
</style>
