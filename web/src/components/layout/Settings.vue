<template>
  <Panel>
    <div class="mb-4 flex flex-col justify-center border-b pb-4 dark:border-wp-background-100">
      <div class="flex items-center justify-between">
        <h1 class="flex items-center gap-1 text-xl text-wp-text-100">
          {{ title }}
          <DocsLink v-if="docsUrl" :topic="title" :url="docsUrl" />
        </h1>
        <slot v-if="$slots.titleActions" name="titleActions" />
      </div>

      <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
        <p v-if="description" class="text-sm text-wp-text-alt-100">{{ description }}</p>
        <div v-if="$slots.headerActions">
          <slot name="headerActions" />
        </div>
      </div>
      <slot name="headerEnd" />
    </div>

    <slot />
  </Panel>
</template>

<script setup lang="ts">
import DocsLink from '~/components/atomic/DocsLink.vue';
import Panel from '~/components/layout/Panel.vue';

defineProps<{
  title: string;
  description?: string;
  docsUrl?: string;
}>();
</script>
