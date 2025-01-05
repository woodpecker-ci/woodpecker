<template>
  <div class="mt-2 flex flex-wrap md:gap-4">
    <router-link
      v-for="tab in tabs"
      :key="tab.title"
      v-slot="{ isActive, isExactActive }"
      :to="tab.to"
      class="text-wp-text-100 flex w-full cursor-pointer items-center border-transparent py-1 md:w-auto md:border-b-2"
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
        class="dark:hover:bg-wp-background-100 hover:bg-wp-background-200 flex w-full min-w-20 flex-row items-center gap-2 rounded-md px-2 py-1 md:justify-center"
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
