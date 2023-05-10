import { defineStore } from 'pinia';
import { computed, reactive, Ref, ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Pipeline, PipelineFeed, PipelineStep } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';
import { comparePipelines, isPipelineActive } from '~/utils/helpers';

export const usePipelineStore = defineStore('pipelines', () => {
  const apiClient = useApiClient();
  const repoStore = useRepoStore();

  const pipelines: Map<number, Map<number, Pipeline>> = reactive(new Map());

  function setPipeline(repoId: number, pipeline: Pipeline) {
    const repoPipelines = pipelines.get(repoId) || new Map();
    repoPipelines.set(pipeline.number, {
      ...(repoPipelines.get(pipeline.number) || {}),
      ...pipeline,
    });
    pipelines.set(repoId, repoPipelines);
  }

  function getRepoPipelines(repoId: Ref<number>) {
    return computed(() => Array.from(pipelines.get(repoId.value)?.values() || []).sort(comparePipelines));
  }

  function getPipeline(repoId: Ref<number>, _pipelineNumber: Ref<string>) {
    return computed(() => {
      const pipelineNumber = parseInt(_pipelineNumber.value, 10);
      return pipelines.get(repoId.value)?.get(pipelineNumber);
    });
  }

  function setStep(repoId: number, pipelineNumber: number, step: PipelineStep) {
    const pipeline = getPipeline(ref(repoId), ref(pipelineNumber.toString())).value;
    if (!pipeline) {
      throw new Error("Can't find pipeline");
    }

    if (!pipeline.steps) {
      pipeline.steps = [];
    }

    pipeline.steps = [...pipeline.steps.filter((p) => p.pid !== step.pid), step];
    setPipeline(repoId, pipeline);
  }

  async function loadRepoPipelines(repoId: number) {
    const _pipelines = await apiClient.getPipelineList(repoId);
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
      .sort(comparePipelines)
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
    setStep,
    getRepoPipelines,
    getPipeline,
    loadRepoPipelines,
    loadPipeline,
    activePipelines,
    pipelineFeed,
    loadPipelineFeed,
  };
});
