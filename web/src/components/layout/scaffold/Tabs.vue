<template>
  <div class="flex flex-wrap">
    <router-link
      v-for="tab in tabs"
      :key="tab.title"
      v-slot="{ isActive, isExactActive }"
      :to="tab.to"
      class="border-transparent w-full py-1 md:py-2 md:w-auto md:px-6 flex cursor-pointer md:border-b-2 text-wp-text-100 hover:text-wp-text-200 items-center"
      :active-class="tab.matchChildren ? '!border-wp-text-100' : ''"
      :exact-active-class="tab.matchChildren ? '' : '!border-wp-text-100'"
    >
      <Icon v-if="isExactActive || (isActive && tab.matchChildren)" name="chevron-right" class="md:hidden" />
      <Icon v-else name="blank" class="md:hidden" />
      <span class="flex gap-2 items-center flex-row-reverse md:flex-row">
        <Icon v-if="tab.icon" :name="tab.icon" :class="tab.iconClass" />
        <span>{{ tab.title }}</span>
      </span>
    </router-link>
  </div>
</template>

<script setup lang="ts">
import Icon from '~/components/atomic/Icon.vue';
import { useTabsClient } from '~/compositions/useTabs';

const { tabs } = useTabsClient();
</script>
