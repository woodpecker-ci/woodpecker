<template>
  <a
    :href="`${docsUrl}`"
    :title="$t('documentation_for', { topic: topic })"
    target="_blank"
    class="text-wp-link-100 hover:text-wp-link-200 cursor-pointer mt-1"
    ><Icon name="question" class="!w-4 !h-4"
  /></a>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import useConfig from '~/compositions/useConfig';

const props = defineProps<{
  url: string;
  topic: string;
}>();

const url = toRef(props, 'url');
const topic = toRef(props, 'topic');
const docsUrl = computed(() => (url.value.startsWith('http') ? url.value : `https://woodpecker-ci.org/${url.value}`));
</script>
