<template>
  <button
    type="button"
    class="relative flex items-center py-1 px-2 rounded-md border shadow-sm cursor-pointer transition-all duration-150 overflow-hidden disabled:opacity-50 disabled:cursor-not-allowed"
    :class="{
      'bg-white hover:bg-gray-200 border-gray-300 text-color dark:bg-dark-gray-600 dark:border-dark-400 dark:hover:bg-dark-gray-800':
        color === 'gray',
      'bg-woodpecker-400 hover:bg-woodpecker-500 border-woodpecker-600 text-white dark:text-gray-300 dark:bg-woodpecker-500 dark:hover:bg-woodpecker-400':
        color === 'green',
      'bg-cyan-600 hover:bg-cyan-700 border-cyan-800 text-white dark:text-gray-300 dark:bg-cyan-900 dark:hover:bg-cyan-800':
        color === 'blue',
      'bg-red-500 hover:bg-red-600 border-red-700 text-white dark:text-gray-300 dark:bg-red-900 dark:hover:bg-red-800':
        color === 'red',
      ...passedClasses,
    }"
    :title="title"
    :disabled="disabled"
    @click="doClick"
  >
    <slot>
      <Icon v-if="startIcon" :name="startIcon" class="!w-6 !h-6" :class="{ invisible: isLoading, 'mr-1': text }" />
      <span :class="{ invisible: isLoading }">{{ text }}</span>
      <Icon v-if="endIcon" :name="endIcon" class="ml-2 w-6 h-6" :class="{ invisible: isLoading }" />
      <div
        class="absolute left-0 top-0 right-0 bottom-0 flex items-center justify-center"
        :class="{
          'opacity-100': isLoading,
          'opacity-0': !isLoading,
          'bg-white dark:bg-dark-gray-700': color === 'gray',
          'bg-lime-700': color === 'green',
          'bg-cyan-700': color === 'blue',
          'bg-red-600': color === 'red',
        }"
      >
        <Icon name="loading" class="animate-spin" />
      </div>
    </slot>
  </button>
</template>

<script lang="ts">
import { computed, defineComponent, PropType } from 'vue';
import { RouteLocationRaw, useRouter } from 'vue-router';

import Icon, { IconNames } from '~/components/atomic/Icon.vue';

export default defineComponent({
  name: 'Button',

  components: { Icon },

  props: {
    text: {
      type: String,
      default: null,
    },

    title: {
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

    startIcon: {
      type: String as PropType<IconNames | null>,
      default: null,
    },

    endIcon: {
      type: String as PropType<IconNames | null>,
      default: null,
    },

    isLoading: {
      type: Boolean,
    },
  },

  setup(props, { attrs }) {
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

    const passedClasses = computed(() => {
      const classes: Record<string, boolean> = {};
      const origClass = (attrs.class as string) || '';
      origClass.split(' ').forEach((c) => {
        classes[c] = true;
      });
      return classes;
    });

    return { doClick, passedClasses };
  },
});
</script>
