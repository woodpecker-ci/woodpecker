<template>
  <div
    class="bg-wp-background-100 dark:bg-wp-background-200 border-wp-background-400 w-full overflow-hidden rounded-md border shadow"
  >
    <component
      :is="collapsable ? 'button' : 'div'"
      v-if="title"
      type="button"
      class="text-wp-text-100 bg-wp-background-300 flex w-full gap-2 px-4 py-2 font-bold"
      @click="_collapsed = !_collapsed"
    >
      <Icon
        v-if="collapsable"
        name="chevron-right"
        class="h-6 min-w-6 transition-transform duration-150"
        :class="{ 'rotate-90 transform': !collapsed }"
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
      <div class="text-wp-text-100 w-full p-4">
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
