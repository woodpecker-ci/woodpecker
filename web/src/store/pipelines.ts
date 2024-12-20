import { defineStore } from 'pinia';
import { computed, reactive, ref, type Ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import type { Pipeline, PipelineFeed, PipelineWorkflow } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';
import { comparePipelines, comparePipelinesWithStatus, isPipelineActive } from '~/utils/helpers';

export const usePipelineStore = defineStore('pipelines', () => {
  const apiClient = useApiClient();
  const repoStore = useRepoStore();

  const pipelines: Map<number, Map<number, Pipeline>> = reactive(new Map());

  function setPipeline(repoId: number, pipeline: Pipeline) {
    const repoPipelines = pipelines.get(repoId) || new Map<number, Pipeline>();
    repoPipelines.set(pipeline.number, {
      ...(repoPipelines.get(pipeline.number) || {}),
      ...pipeline,
    });

    // Update last pipeline number for the repo
    const repo = repoStore.repos.get(repoId);
    if (repo?.last_pipeline !== undefined && repo.last_pipeline < pipeline.number) {
      repo.last_pipeline = pipeline.number;
      repoStore.setRepo(repo);
    }

    pipelines.set(repoId, repoPipelines);
  }

  function getRepoPipelines(repoId: Ref<number>) {
    return computed(() => Array.from(pipelines.get(repoId.value)?.values() || []).sort(comparePipelines));
  }

  function getPipeline(repoId: Ref<number>, _pipelineNumber: Ref<string | number>) {
    return computed(() => {
      if (typeof _pipelineNumber.value === 'string') {
        const pipelineNumber = Number.parseInt(_pipelineNumber.value, 10);
        return pipelines.get(repoId.value)?.get(pipelineNumber);
      }

      return pipelines.get(repoId.value)?.get(_pipelineNumber.value);
    });
  }

  function setWorkflow(repoId: number, pipelineNumber: number, workflow: PipelineWorkflow) {
    const pipeline = getPipeline(ref(repoId), ref(pipelineNumber.toString())).value;
    if (!pipeline) {
      throw new Error("Can't find pipeline");
    }

    if (!pipeline.workflows) {
      pipeline.workflows = [];
    }

    pipeline.workflows = [...pipeline.workflows.filter((p) => p.pid !== workflow.pid), workflow];
    setPipeline(repoId, pipeline);
  }

  async function loadRepoPipelines(repoId: number, page?: number) {
    const _pipelines = await apiClient.getPipelineList(repoId, { page });
    _pipelines.forEach((pipeline) => {
      setPipeline(repoId, pipeline);
    });
  }

  async function loadPipeline(repoId: number, pipelinesNumber: number) {
    const pipeline = await apiClient.getPipeline(repoId, pipelinesNumber);
    setPipeline(repoId, pipeline);
  }

  const pipelineFeed = computed(() =>
    Array.from(pipelines.entries())
      .reduce<PipelineFeed[]>((acc, [_repoId, repoPipelines]) => {
        const repoPipelinesArray = Array.from(repoPipelines.entries()).map(
          ([_pipelineNumber, pipeline]) =>
            <PipelineFeed>{
              ...pipeline,
              repo_id: _repoId,
              number: _pipelineNumber,
            },
        );
        return [...acc, ...repoPipelinesArray];
      }, [])
      .sort(comparePipelinesWithStatus)
      .filter((pipeline) => repoStore.ownedRepoIds.includes(pipeline.repo_id)),
  );

  const activePipelines = computed(() => pipelineFeed.value.filter(isPipelineActive));

  async function loadPipelineFeed() {
    await repoStore.loadRepos();

    const _pipelines = await apiClient.getPipelineFeed();
    _pipelines.forEach((pipeline) => {
      setPipeline(pipeline.repo_id, pipeline);
    });
  }

  return {
    pipelines,
    setPipeline,
    setWorkflow,
    getRepoPipelines,
    getPipeline,
    loadRepoPipelines,
    loadPipeline,
    activePipelines,
    pipelineFeed,
    loadPipelineFeed,
  };
});
