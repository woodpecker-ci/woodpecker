<template>
  <Panel>
    <div class="dark:border-wp-background-100 mb-4 flex flex-col justify-center border-b pb-4">
      <div class="flex items-center justify-between">
        <h1 class="text-wp-text-100 flex items-center gap-1 text-xl">
          {{ title }}
          <DocsLink v-if="docsUrl" :topic="title" :url="docsUrl" />
        </h1>
        <slot v-if="$slots.titleActions" name="titleActions" />
      </div>

      <div class="flex flex-wrap items-center justify-between gap-x-4 gap-y-2">
        <p v-if="description" class="text-wp-text-alt-100 text-sm">{{ description }}</p>
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
