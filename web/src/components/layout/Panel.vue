<template>
  <div
    class="border-wp-background-400 bg-wp-background-100 dark:bg-wp-background-200 shadow border rounded-md w-full overflow-hidden"
  >
    <component
      :is="collapsable ? 'button' : 'div'"
      v-if="title"
      type="button"
      class="flex gap-2 bg-wp-background-400 px-4 py-2 w-full font-bold text-wp-text-100"
      @click="_collapsed = !_collapsed"
    >
      <Icon
        v-if="collapsable"
        name="chevron-right"
        class="min-w-6 h-6 transition-transform duration-150"
        :class="{ 'transform rotate-90': !collapsed }"
      />
      {{ title }}
    </component>
    <div
      :class="{
        'max-h-auto': !collapsed,
        'max-h-0': collapsed,
      }"
      class="transition-height overflow-hidden duration-150"
    >
      <div class="p-4 w-full text-wp-text-100">
        <slot />
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';

import Icon from '~/components/atomic/Icon.vue';

const props = defineProps<{
  title?: string;
  collapsable?: boolean;
  collapsedByDefault?: boolean;
}>();

/**
 * _collapsed is used to store the internal state of the panel, but is
 * ignored if the panel is not collapsable.
 */
const _collapsed = ref(props.collapsedByDefault || false);

const collapsed = computed(() => props.collapsable && _collapsed.value);
</script>
