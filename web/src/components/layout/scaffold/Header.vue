<template>
  <div class="bg-white dark:bg-dark-gray-900 border-b dark:border-gray-700">
    <FluidContainer class="!py-0">
      <div
        class="items-center justify-between py-4"
        :class="{
          'grid grid-cols-[1fr,auto,1fr]': searchBoxPresent,
          'flex flex-wrap': !searchBoxPresent,
        }"
      >
        <div class="flex flex-wrap items-center justify-start">
          <IconButton v-if="goBack" icon="back" :title="$t('back')" class="mr-2" @click="goBack" />
          <h1 class="flex flex-wrap text-xl text-color items-center gap-x-2">
            <slot name="title" />
          </h1>
        </div>
        <TextField
          v-if="searchBoxPresent"
          class="w-auto !bg-gray-100 !dark:bg-dark-gray-600"
          :placeholder="$t('search')"
          :model-value="search"
          @update:model-value="(value: string) => $emit('update:search', value)"
        />
        <div class="flex flex-wrap items-center justify-end gap-x-2">
          <slot name="titleActions" />
        </div>
      </div>

      <div v-if="enableTabs" class="flex flex-wrap justify-between">
        <Tabs />
        <div class="flex items-center justify-end gap-x-2 mb-2">
          <slot name="tabActions" />
        </div>
      </div>
    </FluidContainer>
  </div>
</template>

<script setup lang="ts">
import FluidContainer from '~/components/layout/FluidContainer.vue';

import Tabs from './Tabs.vue';

export interface Props {
  goBack?: () => void;
  enableTabs?: boolean;
  search?: string;
}

const props = defineProps<Props>();
defineEmits(['update:search']);

const searchBoxPresent = props.search !== undefined;
</script>
