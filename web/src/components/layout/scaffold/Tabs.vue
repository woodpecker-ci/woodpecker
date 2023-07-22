<template>
  <div class="flex flex-wrap">
    <button
      v-for="tab in tabs"
      :key="tab.id"
      class="w-full py-2 md:w-auto md:py-2 md:px-8 flex cursor-pointer md:border-b-2 text-wp-text-100 hover:text-wp-gray-700 dark:hover:text-wp-gray-400 items-center"
      :class="{
        'border-wp-text-100': activeTab === tab.id,
        'border-transparent': activeTab !== tab.id,
      }"
      type="button"
      @click="selectTab(tab)"
    >
      <Icon v-if="activeTab === tab.id" name="chevron-right" class="md:hidden" />
      <Icon v-else name="blank" class="md:hidden" />
      <span>{{ tab.title }}</span>
    </button>
  </div>
</template>

<script setup lang="ts">
import { useRoute, useRouter } from 'vue-router';

import { Tab, useTabsClient } from '~/compositions/useTabs';

const router = useRouter();
const route = useRoute();

const { activeTab, tabs, disableHashMode } = useTabsClient();

async function selectTab(tab: Tab) {
  if (tab.id === undefined) {
    return;
  }

  activeTab.value = tab.id;
  if (!disableHashMode.value) {
    await router.replace({ params: route.params, hash: `#${tab.id}` });
  }
}
</script>
