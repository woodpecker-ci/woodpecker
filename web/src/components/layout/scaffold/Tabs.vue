<template>
  <div class="flex flex-wrap">
    <button
      v-for="tab in tabs"
      :key="tab.to"
      class="w-full py-1 md:py-2 md:w-auto md:px-6 flex cursor-pointer md:border-b-2 text-wp-text-100 hover:text-wp-text-200 items-center"
      :class="{
        'border-wp-text-100': activeTab === tab.to,
        'border-transparent': activeTab !== tab.to,
      }"
      type="button"
      @click="selectTab(tab)"
    >
      <Icon v-if="activeTab === tab.to" name="chevron-right" class="md:hidden" />
      <Icon v-else name="blank" class="md:hidden" />
      <span class="flex gap-2 items-center flex-row-reverse md:flex-row">
        <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" />
        <span>{{ tab.title }}</span>
      </span>
    </button>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/atomic/Icon.vue';
import { useTabsClient, type Tab } from '~/compositions/useTabs';

const { activeTab, tabs } = useTabsClient();

async function selectTab(tab: Tab) {
  if (tab.to === undefined) {
    return;
  }

  activeTab.value = tab.to;
}
</script>
