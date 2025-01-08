<template>
  <header
    class="border-b border-wp-background-400 bg-wp-background-100 text-wp-text-100 dark:border-wp-background-100 dark:bg-wp-background-300"
    :class="{ 'md:px-4': fullWidth }"
  >
    <Container :full-width="fullWidth" class="!py-0">
      <div class="flex w-full flex-col gap-2 py-3 md:flex-row md:items-center md:justify-between md:gap-10">
        <div
          class="flex min-h-10 content-start items-center"
          :class="{
            'md:flex-1': searchBoxPresent,
          }"
        >
          <IconButton
            v-if="goBack"
            icon="back"
            :title="$t('back')"
            class="md:display-unset mr-2 hidden h-8 w-8 flex-shrink-0 md:justify-between"
            @click="goBack"
          />
          <h1 class="flex min-w-0 items-center gap-x-2 text-xl text-wp-text-100">
            <slot name="title" />
          </h1>
        </div>
        <TextField
          v-if="searchBoxPresent"
          class="order-3 w-full flex-grow md:order-none md:w-auto"
          :aria-label="$t('search')"
          :placeholder="$t('search')"
          :model-value="search"
          @update:model-value="(value: string) => $emit('update:search', value)"
        />
        <div
          v-if="$slots.headerActions"
          class="flex min-w-0 items-center gap-x-2 md:justify-end"
          :class="{
            'md:flex-1': searchBoxPresent,
          }"
        >
          <slot name="headerActions" />
        </div>
      </div>

      <div v-if="enableTabs" class="flex flex-row items-center justify-between py-0">
        <Tabs />
        <div v-if="$slots.headerActions" class="flex content-start justify-end">
          <slot name="tabActions" />
        </div>
      </div>
    </Container>
  </header>
</template>

<script setup lang="ts">
import { computed } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import TextField from '~/components/form/TextField.vue';
import Container from '~/components/layout/Container.vue';

import Tabs from './Tabs.vue';

const props = defineProps<{
  goBack?: () => void;
  enableTabs?: boolean;
  search?: string;
  fullWidth?: boolean;
}>();

defineEmits<{
  (event: 'update:search', query: string): void;
}>();

const searchBoxPresent = computed(() => props.search !== undefined);
</script>
