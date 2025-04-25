<template>
  <div class="mb-4 flex w-full justify-center">
    <span class="text-wp-text-100 text-xl">{{ $t('repo.pipeline.pipelines_for_pr', { index: pullRequest }) }}</span>
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
  pullRequest: string;
}>();
const pullRequest = toRef(props, 'pullRequest');
const repo = requiredInject('repo');
if (!repo.value.pr_enabled || !repo.value.allow_pr) {
  throw new Error('Unexpected: pull requests are disabled for repo');
}

const allPipelines = requiredInject('pipelines');
const pipelines = computed(() =>
  allPipelines.value.filter(
    (b) =>
      (b.event === 'pull_request' || b.event === 'pull_request_closed') &&
      b.ref
        .replaceAll('refs/pull/', '')
        .replaceAll('refs/merge-requests/', '')
        .replaceAll('refs/pull-requests/', '')
        .replaceAll('/from', '')
        .replaceAll('/merge', '')
        .replaceAll('/head', '') === pullRequest.value,
  ),
);

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.activity'), pullRequest.value, repo.value.full_name]));
</script>
