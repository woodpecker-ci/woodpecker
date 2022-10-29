<template>
  <div class="flex flex-col mt-2 mb-4">
    <div class="flex items-center text-color font-bold mb-2">
      <label v-bind="$attrs">{{ label }}</label>
      <DocsLink v-if="docsUrl" :topic="label" :url="docsUrl" class="ml-2" />
    </div>
    <slot />
    <div v-if="$slots['description']" class="ml-1 text-color-alt">
      <slot name="description" />
    </div>
  </div>
</template>

<script lang="ts" setup>
import DocsLink from '~/components/atomic/DocsLink.vue';

defineProps<{
  label: string;
  docsUrl?: string;
}>();
</script>

<script lang="ts">
export default {
  inheritAttrs: false,
};
</script>
