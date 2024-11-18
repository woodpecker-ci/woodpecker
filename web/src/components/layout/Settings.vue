<template>
  <Panel>
    <div class="flex flex-col border-b mb-4 pb-4 justify-center dark:border-wp-background-100">
      <h1 class="text-xl text-wp-text-100 flex items-center gap-1">
        {{ title }}
        <DocsLink v-if="docsUrl" :topic="title" :url="docsUrl" />
      </h1>

      <div class="flex flex-wrap gap-x-4 gap-y-2 items-center justify-between">
        <p v-if="desc" class="text-sm text-wp-text-alt-100">{{ desc }}</p>
        <div v-if="$slots.titleActions">
          <slot name="titleActions" />
        </div>
      </div>

      <Warning v-if="warning" class="text-sm mt-4" :text="warning" />
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
