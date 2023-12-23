<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-wp-background-100">
      <div class="ml-2">
        <h1 class="text-xl text-wp-text-100">{{ title }}</h1>
        <p v-if="desc" class="text-sm text-wp-text-alt-100">
          {{ desc }}
          <DocsLink v-if="docsUrl" :topic="title" :url="docsUrl" />
        </p>
        <Warning v-if="warning" class="text-sm mt-1" :text="warning" />
      </div>

      <div class="ml-auto">
        <slot v-if="$slots.titleActions" name="titleActions" />
      </div>
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
