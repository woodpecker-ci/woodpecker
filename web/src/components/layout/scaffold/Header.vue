<template>
  <header class="bg-white dark:bg-dark-gray-900 border-b dark:border-gray-700 text-color">
    <FluidContainer class="!py-0">
      <div class="flex w-full items-center justify-between py-4 <md:flex-row <md:gap-y-4">
        <div
          class="flex items-center min-w-0 justify-start <md:justify-center"
          :class="{
            'md:flex-1': searchBoxPresent,
          }"
        >
          <IconButton
            v-if="goBack"
            icon="back"
            :title="$t('back')"
            class="flex-shrink-0 mr-2 <md:hidden"
            @click="goBack"
          />
          <h1 class="flex text-xl min-w-0 text-color items-center gap-x-2">
            <slot name="title" />
          </h1>
        </div>
        <TextField
          v-if="searchBoxPresent"
          class="w-auto !bg-gray-100 !dark:bg-dark-gray-600 placeholder-gray-500 <md:w-full <md:order-3"
          :placeholder="$t('search')"
          :model-value="search"
          @update:model-value="(value: string) => $emit('update:search', value)"
        />
        <div
          v-if="$slots.titleActions"
          class="flex items-center justify-end gap-x-2 <md:w-full <md:justify-center"
          :class="{
            'md:flex-1': searchBoxPresent,
          }"
        >
          <slot name="titleActions" />
        </div>
      </div>

      <div v-if="enableTabs" class="flex justify-between">
        <Tabs class="<md:order-2" />
        <div
          v-if="$slots.titleActions"
          class="flex items-center justify-end gap-x-2 md:mb-2 <md:w-full <md:justify-center <md:order-1"
        >
          <slot name="tabActions" />
        </div>
      </div>
    </FluidContainer>
  </header>
</template>

<script setup lang="ts">
import TextField from '~/components/form/TextField.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';

import Tabs from './Tabs.vue';

const props = defineProps<{
  goBack?: () => void;
  enableTabs?: boolean;
  search?: string;
}>();
defineEmits(['update:search']);

const searchBoxPresent = props.search !== undefined;
</script>
