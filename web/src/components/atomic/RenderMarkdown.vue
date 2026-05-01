<template>
  <span v-html="contentHTML" />
</template>

<script setup lang="ts">
import DOMPurify from 'dompurify';
import { marked } from 'marked';
import { computed } from 'vue';

const props = defineProps<{
  content: string;
  inline?: boolean;
}>();

const contentHTML = computed<string>(() => {
  const dirtyHTML = props.inline ? marked.parseInline(props.content) : marked.parse(props.content);
  return DOMPurify.sanitize(dirtyHTML as string, { USE_PROFILES: { html: true } });
});
</script>
