<template>
  <span v-html="marked(props.content)" />
</template>

<script setup lang="ts">
import { marked } from 'marked';

const props = defineProps<{
  content: string;
}>();
</script>
