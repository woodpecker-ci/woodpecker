<template>
  <header
    class="bg-wp-background-100 border-b-1 border-wp-background-400 dark:border-wp-background-100 dark:bg-wp-background-300 text-wp-text-100"
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
            class="<md:hidden mr-2 h-8 w-8 flex-shrink-0 md:justify-between"
            @click="goBack"
          />
          <h1 class="text-wp-text-100 flex min-w-0 items-center gap-x-2 text-xl">
            <slot name="title" />
          </h1>
        </div>
        <TextField
          v-if="searchBoxPresent"
          class="<md:w-full <md:order-3 w-auto flex-grow"
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

      <div v-if="enableTabs" class="flex flex-col py-2 md:flex-row md:items-center md:justify-between md:py-0">
        <Tabs class="<md:order-2" />
        <div v-if="$slots.headerActions" class="flex content-start md:justify-end">
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
