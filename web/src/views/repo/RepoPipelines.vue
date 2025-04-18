<template>
  <PipelineList :pipelines="pipelines" :repo="repo" />
</template>

<script lang="ts" setup>
import { computed, inject } from 'vue';
import type { Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Pipeline, Repo, RepoPermissions } from '~/lib/api/types';

const repo = inject<Ref<Repo>>('repo');
const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
if (!repo || !repoPermissions) {
  throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
}

const pipelines = inject<Ref<Pipeline[]>>('pipelines');

const { t } = useI18n()
useWPTitle(computed(() => [t('repo.activity'), repo.value.name]));
</script>
