<template>
  <div class="flex flex-wrap md:gap-4 mt-2">
    <router-link
      v-for="tab in tabs"
      :key="tab.title"
      :to="tab.to"
      class="flex items-center !border-wp-text-100 py-1 border-transparent md:border-b-2 w-full md:w-auto text-wp-text-100 cursor-pointer"
      active-class="underline !border-wp-text-100"
      exact-active-class="underline !border-wp-text-100"
    >
      <span
        class="flex flex-row md:justify-center items-center gap-2 dark:hover:bg-wp-background-100 hover:bg-wp-background-200 py-1 sm:pr-2 sm:pl-2.5 md:pl-2 rounded-md w-full min-w-20"
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
