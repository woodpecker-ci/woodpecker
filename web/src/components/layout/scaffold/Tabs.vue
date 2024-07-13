<template>
  <div class="flex flex-wrap">
    <button
      v-for="tab in tabs"
      :key="tab.id"
      class="w-full py-1 md:py-2 md:w-auto md:px-6 flex cursor-pointer md:border-b-2 text-wp-text-100 hover:text-wp-text-200 items-center"
      :class="{
        'border-wp-text-100': activeTab === tab.id,
        'border-transparent': activeTab !== tab.id,
      }"
      type="button"
      @click="selectTab(tab)"
    >
      <Icon v-if="activeTab === tab.id" name="chevron-right" class="md:hidden" />
      <Icon v-else name="blank" class="md:hidden" />
      <span class="flex gap-2 items-center flex-row-reverse md:flex-row">
        <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" />
        <span>{{ tab.title }}</span>
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';

import { useTabsClient, type Tab } from '~/compositions/useTabs';

const router = useRouter();
const route = useRoute();

const { activeTab, tabs, disableUrlHashMode } = useTabsClient();

async function selectTab(tab: Tab) {
  if (tab.id === undefined) {
    return;
  }

  activeTab.value = tab.id;

  if (!disableUrlHashMode.value) {
    await router.replace({ params: route.params, hash: `#${tab.id}` });
  }
}
</script>
