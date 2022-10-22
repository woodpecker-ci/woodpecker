<template>
  <div class="bg-white dark:bg-dark-gray-900 border-b dark:border-gray-700">
    <FluidContainer class="!py-0">
      <Header :go-back="goBack">
        <template #title><slot name="headerTitle" /></template>
        <template v-if="!!$slots.headerCenterBox" #centerBox><slot name="headerCenterBox" /></template>
        <template #actions><slot name="headerActions" /></template>
      </Header>

      <div v-if="enableTabs" class="flex flex-wrap justify-between">
        <Tabs />
        <div class="flex items-center justify-end gap-x-2 mb-2">
          <slot name="tabActions" />
        </div>
      </div>
    </FluidContainer>
  </div>
  <FluidContainer>
    <slot />
  </FluidContainer>
</template>

<script setup lang="ts">
import { toRef } from 'vue';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import { useTabsProvider } from '~/compositions/useTabs';

import Header from './Header.vue';
import Tabs from './Tabs.vue';

export interface Props {
  // Header
  goBack?: () => void;

  // Tabs
  enableTabs?: boolean;
  disableHashMode?: boolean;
  activeTab?: string;
}

const props = withDefaults(defineProps<Props>(), {
  goBack: undefined,
  // eslint-disable-next-line vue/no-boolean-default
  disableHashMode: false,
  // eslint-disable-next-line vue/no-boolean-default
  enableTabs: false,
  activeTab: '',
});

const emit = defineEmits(['update:activeTab']);

if (props.enableTabs) {
  useTabsProvider({
    activeTabProp: toRef(props, 'activeTab'),
    disableHashMode: toRef(props, 'disableHashMode'),
    updateActiveTabProp: (value) => emit('update:activeTab', value),
  });
}
</script>
