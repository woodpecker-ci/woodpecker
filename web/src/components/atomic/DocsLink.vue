<template>
  <a
    :href="`${docsUrl}`"
    :title="$t('documentation_for', { topic: topic })"
    target="_blank"
    class="text-blue-500 hover:text-blue-600 cursor-pointer mt-1"
    ><Icon name="question" class="!w-4 !h-4"
  /></a>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';

const props = defineProps<{
  url: string;
  topic: string;
}>();

const docsBaseUrl = window.WOODPECKER_DOCS;
const url = toRef(props, 'url');
const topic = toRef(props, 'topic');
const docsUrl = computed(() => (url.value.startsWith('http') ? url.value : `${docsBaseUrl}${url.value}`));
</script>
