<template>
  <div class="flex flex-wrap mt-2 md:gap-4">
    <router-link
      v-for="tab in tabs"
      :key="tab.title"
      v-slot="{ isActive, isExactActive }"
      :to="tab.to"
      class="border-transparent w-full py-1 md:w-auto flex cursor-pointer md:border-b-2 text-wp-text-100 items-center"
      :active-class="tab.matchChildren ? '!border-wp-text-100' : ''"
      :exact-active-class="tab.matchChildren ? '' : '!border-wp-text-100'"
    >
      <Icon
        v-if="isExactActive || (isActive && tab.matchChildren)"
        name="chevron-right"
        class="md:hidden flex-shrink-0"
        size="20"
      />
      <Icon v-else name="blank" class="md:hidden" />
      <span
        class="flex gap-2 items-center md:justify-center flex-row py-1 px-2 w-full min-w-20 dark:hover:bg-wp-background-100 hover:bg-wp-background-200 rounded-md"
      >
        <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" class="flex-shrink-0" size="20" />
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
