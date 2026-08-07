<template>
  <div class="mb-4 flex w-full justify-center">
    <span class="text-wp-text-100 text-xl">{{ $t('repo.pipeline.pipelines_for_tag', { tag }) }}</span>
  </div>
  <PipelineList :pipelines="pipelines" :repo="repo" />
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';

const props = defineProps<{
  tag: string;
}>();

const tag = toRef(props, 'tag');
const repo = requiredInject('repo');

const allPipelines = requiredInject('pipelines');
const pipelines = computed(() =>
  allPipelines.value.filter((p) => p.tag_title === tag.value || p.ref === `refs/tags/${tag.value}`),
);

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.activity'), tag.value, repo.value.full_name]));
</script>
