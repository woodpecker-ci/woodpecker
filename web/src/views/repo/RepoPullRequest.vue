<template>
  <div class="flex w-full mb-4 justify-center">
    <span class="text-wp-text-100 text-xl">{{ $t('repo.pipeline.pipelines_for_pr', { index: pullRequest }) }}</span>
  </div>
  <PipelineList :pipelines="pipelines" :repo="repo" />
</template>

<script lang="ts" setup>
import { computed, inject, Ref, toRef } from 'vue';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { Pipeline, Repo, RepoPermissions } from '~/lib/api/types';

const props = defineProps<{
  pullRequest: string;
}>();
const pullRequest = toRef(props, 'pullRequest');
const repo = inject<Ref<Repo>>('repo');
const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
if (!repo || !repoPermissions) {
  throw new Error('Unexpected: "repo" and "repoPermissions" should be provided at this place');
}

const allPipelines = inject<Ref<Pipeline[]>>('pipelines');
const pipelines = computed(() =>
  allPipelines?.value.filter(
    (b) =>
      b.event === 'pull_request' &&
      b.ref
        .replaceAll('refs/pull/', '')
        .replaceAll('refs/merge-requests/', '')
        .replaceAll('/merge', '')
        .replaceAll('/head', '') === pullRequest.value,
  ),
);
</script>
