<template>
  <PipelineList :pipelines="pipelines" :repo="repo" />
</template>

<script lang="ts" setup>
import { inject, Ref } from 'vue';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { Pipeline, Repo, RepoPermissions } from '~/lib/api/types';

const repo = inject<Ref<Repo>>('repo');
const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
if (!repo || !repoPermissions) {
  throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
}

const pipelines = inject<Ref<Pipeline[]>>('pipelines');
</script>
