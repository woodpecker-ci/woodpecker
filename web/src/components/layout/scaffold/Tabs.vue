<template>
  <div class="flex flex-wrap md:gap-4 mt-2">
    <router-link
      v-for="tab in tabs"
      :key="tab.title"
      :to="tab.to"
      class="flex items-center px-2 py-1 border-transparent md:border-b-2 w-auto md:w-auto text-wp-text-100 cursor-pointer"
      :active-class="tab.matchChildren ? '!border-wp-text-100 md:dark:bg-wp-background-300 dark:bg-wp-background-100 md:bg-wp-background-100 bg-wp-background-200 px-2 w-auto rounded-md hover:px-2' : ''"
      :exact-active-class="tab.matchChildren ? '' : '!border-wp-text-100 md:dark:bg-wp-background-300 dark:bg-wp-background-100 md:bg-wp-background-100 bg-wp-background-200 px-2 w-auto rounded-md hover:px-2'"
    >
      <span
        class="flex flex-row justify-center items-center gap-2 dark:hover:bg-wp-background-100 hover:bg-wp-background-200 py-1 rounded-md w-full min-w-20"
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
