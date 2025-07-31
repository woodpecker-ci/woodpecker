<template>
  <div v-html="formattedCode" />
</template>

<script setup lang="ts">
import { type BundledLanguage } from 'shiki';
import { computed, toRef } from 'vue';

import useSyntaxHighlighter from '~/compositions/useSyntaxHighlighter';

const props = defineProps<{
  code: string;
  language?: BundledLanguage;
}>();

const code = toRef(props.code);
const { formattedCode } = useSyntaxHighlighter(
  code,
  computed(() => props.language ?? 'yaml'),
);
</script>

<style scoped>
::v-deep(.shiki) {
  background-color: transparent !important;
}
</style>
