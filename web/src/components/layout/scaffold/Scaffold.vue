<template>
  <Header
    :go-back="goBack"
    :enable-tabs="enableTabs"
    :search="search"
    :full-width="fullWidth"
    @update:search="(value) => $emit('update:search', value)"
  >
    <template #title><slot name="title" /></template>
    <template v-if="$slots.titleActions" #titleActions><slot name="titleActions" /></template>
    <template v-if="$slots.tabActions" #tabActions><slot name="tabActions" /></template>
  </Header>

  <slot v-if="fluidContent" />
  <FluidContainer v-else>
    <slot />
  </FluidContainer>
</template>

<script setup lang="ts">
import { toRef } from 'vue';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import { useTabsProvider } from '~/compositions/useTabs';

import Header from './Header.vue';

const props = defineProps<{
  // Header
  goBack?: () => void;
  search?: string;

  // Tabs
  enableTabs?: boolean;
  disableHashMode?: boolean;
  activeTab?: string;

  // Content
  fluidContent?: boolean;
  fullWidth?: boolean;
}>();

const emit = defineEmits<{
  (event: 'update:activeTab', value: string): void;
  (event: 'update:search', value: string): void;
}>();

if (props.enableTabs) {
  useTabsProvider({
    activeTabProp: toRef(props, 'activeTab'),
    disableHashMode: toRef(props, 'disableHashMode'),
    updateActiveTabProp: (value) => emit('update:activeTab', value),
  });
}
</script>
