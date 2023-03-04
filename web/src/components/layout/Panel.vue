<template>
  <div class="rounded-md w-full shadow overflow-hidden bg-gray-300 dark:bg-dark-gray-700">
    <component
      :is="collapsable ? 'button' : 'div'"
      v-if="title"
      type="button"
      class="flex w-full font-bold gap-2 text-gray-200 bg-gray-400 dark:bg-dark-gray-800 px-4 py-2"
      @click="collapsed = !collapsed"
    >
      <Icon
        v-if="collapsable"
        name="chevron-right"
        class="transition-transform duration-150 min-w-6 h-6"
        :class="{ 'transform rotate-90': !isCollapsable }"
      />
      {{ title }}
    </component>
    <div
      :class="{
        'max-h-screen': !isCollapsable,
        'max-h-0': isCollapsable,
      }"
      class="transition-height duration-150 overflow-hidden"
    >
      <div class="w-full p-4 bg-white dark:bg-dark-gray-700 text-color">
        <slot />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, ref } from 'vue';

import Icon from '~/components/atomic/Icon.vue';

export default defineComponent({
  name: 'Panel',
  components: { Icon },

  props: {
    title: {
      type: String,
      default: '',
    },

    collapsable: {
      type: Boolean,
    },
  },

  setup(props) {
    const collapsed = ref(false);

    const isCollapsable = computed(() => props.collapsable && collapsed.value);

    return {
      isCollapsable,
      collapsed,
    };
  },
});
</script>
