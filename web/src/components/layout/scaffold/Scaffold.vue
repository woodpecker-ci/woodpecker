<template>
  <Header
    :go-back="goBack"
    :enable-tabs="enableTabs"
    :search="search"
    :full-width="fullWidthHeader"
    @update:search="(value) => $emit('update:search', value)"
  >
    <template #title><slot name="title" /></template>
    <template v-if="$slots.titleActions" #titleActions><slot name="titleActions" /></template>
    <template v-if="$slots.tabActions" #tabActions><slot name="tabActions" /></template>
  </Header>

  <slot v-if="fluidContent" />
  <Container v-else>
    <slot />
  </Container>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue';

import Container from '~/components/layout/Container.vue';
import { useTabsProvider } from '~/compositions/useTabs';

import Header from './Header.vue';

const props = defineProps<{
  // Header
  goBack?: () => void;
  search?: string;
  fullWidthHeader?: boolean;

  // Tabs
  enableTabs?: boolean;
  disableTabUrlHashMode?: boolean;
  activeTab?: string;

  // Content
  fluidContent?: boolean;
}>();

const emit = defineEmits<{
  (event: 'update:activeTab', value: string | undefined): void;
  (event: 'update:search', value: string): void;
}>();

if (props.enableTabs) {
  const internalActiveTab = ref(props.activeTab);

  watch(
    () => props.activeTab,
    (activeTab) => {
      internalActiveTab.value = activeTab;
    },
  );

  useTabsProvider({
    activeTab: computed({
      get: () => internalActiveTab.value,
      set: (value) => {
        internalActiveTab.value = value;
        emit('update:activeTab', value);
      },
    }),
    disableUrlHashMode: computed(() => props.disableTabUrlHashMode || false),
  });
}
</script>
