<template>
  <div class="flex flex-wrap md:gap-4 mt-2">
    <router-link
      v-for="tab in tabs"
      :key="tab.title"
      v-slot="{ isActive, isExactActive }"
      :to="tab.to"
      class="flex items-center py-1 border-transparent md:border-b-2 w-full md:w-auto text-wp-text-100 cursor-pointer"
      :active-class="tab.matchChildren ? '!border-wp-text-100' : ''"
      :exact-active-class="tab.matchChildren ? '' : '!border-wp-text-100'"
    >
      <Icon
        v-if="isExactActive || (isActive && tab.matchChildren)"
        name="chevron-right"
        class="flex-shrink-0 md:hidden"
      />
      <Icon v-else name="blank" class="md:hidden" />
      <span
        class="flex flex-row md:justify-center items-center gap-2 dark:hover:bg-wp-background-100 hover:bg-wp-background-200 px-2 py-1 rounded-md w-full min-w-20"
      >
        <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" class="flex-shrink-0" />
        <span>{{ tab.title }}</span>
        <CountBadge v-if="tab.count" :value="tab.count" />
      </span>
    </router-link>
  </div>
</template>

<script setup lang="ts">
import CountBadge from '~/components/atomic/CountBadge.vue';
import Icon from '~/components/atomic/Icon.vue';
import { useTabsClient } from '~/compositions/useTabs';

const { tabs } = useTabsClient();
</script>
