<template>
  <div class="rounded-md w-full shadow overflow-hidden bg-gray-300 dark:bg-dark-gray-700">
    <component
      :is="collapsable ? 'button' : 'div'"
      v-if="title"
      type="button"
      class="flex w-full font-bold gap-2 text-gray-200 bg-gray-400 dark:bg-dark-gray-800 px-4 py-2"
      @click="collapsed && (_collapsed = !_collapsed)"
    >
      <Icon
        v-if="collapsable"
        name="chevron-right"
        class="transition-transform duration-150 min-w-6 h-6"
        :class="{ 'transform rotate-90': !collapsed }"
      />
      {{ title }}
    </component>
    <div
      :class="{
        'max-h-auto': !collapsed,
        'max-h-0': collapsed,
      }"
      class="transition-height duration-150 overflow-hidden"
    >
      <div class="w-full p-4 bg-white dark:bg-dark-gray-700 text-color">
        <slot />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';

import Icon from '~/components/atomic/Icon.vue';

const props = withDefaults(
  defineProps<{
    title?: string;
    collapsable?: boolean;
  }>(),
  {
    title: '',
  },
);

/**
 * _collapsed is used to store the internal state of the panel, but is
 * ignored if the panel is not collapsable.
 */
const _collapsed = ref(false);

const collapsed = computed(() => props.collapsable && _collapsed.value);
</script>
