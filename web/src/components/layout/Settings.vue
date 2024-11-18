<template>
  <Panel>
    <h1 class="text-xl text-wp-text-100">{{ title }}</h1>
    <div class="flex flex-col gap-4 border-b mb-4 pb-4">
      <div class="flex flex-col sm:flex-row gap-4 sm:gap-12 md:justify-between dark:border-wp-background-100">
        <div v-if="desc" class="flex items-center gap-x-2 text-sm text-wp-text-alt-100">
          <span class="flex flex-grow-0">{{ desc }}</span>
          <DocsLink v-if="docsUrl" class="flex flex-grow-0" :topic="title" :url="docsUrl" />
        </div>

        <div>
          <slot v-if="$slots.titleActions" name="titleActions" />
        </div>
      </div>
      <Warning v-if="warning" class="text-sm mt-1" :text="warning" />
    </div>

    <slot />
  </Panel>
</template>

<script setup lang="ts">
import DocsLink from '~/components/atomic/DocsLink.vue';
import Warning from '~/components/atomic/Warning.vue';
import Panel from '~/components/layout/Panel.vue';

defineProps<{
  title: string;
  desc?: string;
  docsUrl?: string;
  warning?: string;
}>();
</script>
