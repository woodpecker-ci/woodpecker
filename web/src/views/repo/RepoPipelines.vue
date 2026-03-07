<template>
  <PipelineList :pipelines="data" :loading="loading" />
</template>

<script lang="ts" setup>
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Pipeline } from '~/lib/api/types';
import { usePipelineStore } from '~/store/pipelines';

const apiClient = useApiClient();
const repo = requiredInject('repo');
const pipelineStore = usePipelineStore();

async function loadPipelines(page: number): Promise<Pipeline[]> {
  const pipelines = await apiClient.getPipelineList(repo.value.id, { page });
  pipelines.forEach((pipeline) => {
    pipelineStore.setPipeline(repo.value.id, pipeline);
  });
  return pipelines;
}

const { resetPage, data, loading } = usePagination(loadPipelines);

watch(repo, resetPage);

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.activity'), repo.value.full_name]));
</script>
