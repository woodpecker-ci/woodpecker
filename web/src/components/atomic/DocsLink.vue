<template>
  <a
    :href="`${docsUrl}`"
    :title="$t('documentation_for', { topic })"
    target="_blank"
    class="text-wp-link-100 hover:text-wp-link-200 cursor-pointer"
  >
    <Icon name="question" class="!w-5 !h-5" />
  </a>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';

const props = defineProps<{
  url: string;
  topic: string;
}>();

const url = toRef(props, 'url');
const topic = toRef(props, 'topic');
const docsUrl = computed(() => (url.value.startsWith('http') ? url.value : `https://woodpecker-ci.org/${url.value}`));
</script>
