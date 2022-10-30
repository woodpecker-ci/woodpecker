<template>
  <Header
    :go-back="goBack"
    :enable-tabs="enableTabs"
    :search="search"
    @update:search="(value) => $emit('update:search', value)"
  >
    <template #title><slot name="title" /></template>
    <template v-if="$slots.titleActions" #titleActions><slot name="titleActions" /></template>
    <template v-if="$slots.tabActions" #tabActions><slot name="tabActions" /></template>
  </Header>

  <slot v-if="noFluidWrapper" />
  <FluidContainer v-else>
    <slot />
  </FluidContainer>
</template>

<script setup lang="ts">
import { toRef } from 'vue';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import { useTabsProvider } from '~/compositions/useTabs';

import Header from './Header.vue';

export interface Props {
  // Header
  goBack?: () => void;
  search?: string;

  // Tabs
  enableTabs?: boolean;
  disableHashMode?: boolean;
  activeTab: string;

  // Content
  noFluidWrapper?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  goBack: undefined,
  search: undefined,
  // eslint-disable-next-line vue/no-boolean-default
  disableHashMode: false,
  // eslint-disable-next-line vue/no-boolean-default
  enableTabs: false,
  activeTab: '',
  // eslint-disable-next-line vue/no-boolean-default
  noFluidWrapper: false,
});

const emit = defineEmits(['update:activeTab', 'update:search']);

if (props.enableTabs) {
  useTabsProvider({
    activeTabProp: toRef(props, 'activeTab'),
    disableHashMode: toRef(props, 'disableHashMode'),
    updateActiveTabProp: (value) => emit('update:activeTab', value),
  });
}
</script>
