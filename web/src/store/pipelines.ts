import { defineStore } from 'pinia';
import { computed, reactive, Ref, ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Pipeline, PipelineFeed, PipelineWorkflow } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';
import { comparePipelines, isPipelineActive, repoSlug } from '~/utils/helpers';

export const usePipelineStore = defineStore('pipelines', () => {
  const apiClient = useApiClient();
  const repoStore = useRepoStore();

  const pipelines: Map<string, Map<number, Pipeline>> = reactive(new Map());

  function setPipeline(owner: string, repo: string, pipeline: Pipeline) {
    const _repoSlug = repoSlug(owner, repo);
    const repoPipelines = pipelines.get(_repoSlug) || new Map();
    repoPipelines.set(pipeline.number, {
      ...(repoPipelines.get(pipeline.number) || {}),
      ...pipeline,
    });
    pipelines.set(_repoSlug, repoPipelines);
  }

  function getRepoPipelines(owner: Ref<string>, repo: Ref<string>) {
    return computed(() => {
      const slug = repoSlug(owner.value, repo.value);
      return Array.from(pipelines.get(slug)?.values() || []).sort(comparePipelines);
    });
  }

  function getPipeline(owner: Ref<string>, repo: Ref<string>, _pipelineNumber: Ref<string>) {
    return computed(() => {
      const slug = repoSlug(owner.value, repo.value);
      const pipelineNumber = parseInt(_pipelineNumber.value, 10);
      return pipelines.get(slug)?.get(pipelineNumber);
    });
  }

  function setWorkflow(owner: string, repo: string, pipelineNumber: number, workflow: PipelineWorkflow) {
    const pipeline = getPipeline(ref(owner), ref(repo), ref(pipelineNumber.toString())).value;
    if (!pipeline) {
      throw new Error("Can't find pipeline");
    }

    if (!pipeline.workflows) {
      pipeline.workflows = [];
    }

    pipeline.workflows = [...pipeline.workflows.filter((p) => p.pid !== workflow.pid), workflow];
    setPipeline(owner, repo, pipeline);
  }

  async function loadRepoPipelines(owner: string, repo: string) {
    const _pipelines = await apiClient.getPipelineList(owner, repo);
    _pipelines.forEach((pipeline) => {
      setPipeline(owner, repo, pipeline);
    });
  }

  async function loadPipeline(owner: string, repo: string, pipelinesNumber: number) {
    const pipeline = await apiClient.getPipeline(owner, repo, pipelinesNumber);
    setPipeline(owner, repo, pipeline);
  }

  const pipelineFeed = computed(() =>
    Array.from(pipelines.entries())
      .reduce<PipelineFeed[]>((acc, [_repoSlug, repoPipelines]) => {
        const repoPipelinesArray = Array.from(repoPipelines.entries()).map(
          ([_pipelineNumber, pipeline]) =>
            <PipelineFeed>{
              ...pipeline,
              full_name: _repoSlug,
              owner: _repoSlug.split('/')[0],
              name: _repoSlug.split('/')[1],
              number: _pipelineNumber,
            },
        );
        return [...acc, ...repoPipelinesArray];
      }, [])
      .sort(comparePipelines)
      .filter((pipeline) => repoStore.ownedRepoSlugs.includes(pipeline.full_name)),
  );

  const activePipelines = computed(() => pipelineFeed.value.filter(isPipelineActive));

  async function loadPipelineFeed() {
    await repoStore.loadRepos();

    const _pipelines = await apiClient.getPipelineFeed();
    _pipelines.forEach((pipeline) => {
      setPipeline(pipeline.owner, pipeline.name, pipeline);
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
