import { defineStore } from 'pinia';
import { computed, Ref, ref, toRef } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Pipeline, PipelineFeed, PipelineStep } from '~/lib/api/types';
import { comparePipelines, comparePipelinesWithStatus, isPipelineActive, repoSlug } from '~/utils/helpers';

const apiClient = useApiClient();

export default defineStore({
  id: 'pipelines',

  state: () => ({
    pipelines: {} as Record<string, Record<number, Pipeline>>,
    pipelineFeed: [] as PipelineFeed[],
  }),

  getters: {
    sortedPipelineFeed(state) {
      return state.pipelineFeed.sort(comparePipelines);
    },
    activePipelines(state) {
      return state.pipelineFeed.filter(isPipelineActive);
    },
  },

  actions: {
    // setters
    setPipeline(owner: string, repo: string, pipeline: Pipeline) {
      // eslint-disable-next-line @typescript-eslint/naming-convention
      const _repoSlug = repoSlug(owner, repo);
      if (!this.pipelines[_repoSlug]) {
        this.pipelines[_repoSlug] = {};
      }

      const repoPipelines = this.pipelines[_repoSlug];

      // merge with available data
      repoPipelines[pipeline.number] = { ...(repoPipelines[pipeline.number] || {}), ...pipeline };

      this.pipelines = {
        ...this.pipelines,
        [_repoSlug]: repoPipelines,
      };
    },
    setStep(owner: string, repo: string, pipelineNumber: number, step: PipelineStep) {
      const pipeline = this.getPipeline(ref(owner), ref(repo), ref(pipelineNumber.toString())).value;
      if (!pipeline) {
        throw new Error("Can't find pipeline");
      }

      if (!pipeline.steps) {
        pipeline.steps = [];
      }

      pipeline.steps = [...pipeline.steps.filter((p) => p.pid !== step.pid), step];
      this.setPipeline(owner, repo, pipeline);
    },
    setPipelineFeedItem(pipeline: PipelineFeed) {
      const pipelineFeed = this.pipelineFeed.filter((b) => b.id !== pipeline.id);
      this.pipelineFeed = [...pipelineFeed, pipeline];
    },

    // getters
    getPipelines(owner: Ref<string>, repo: Ref<string>) {
      return computed(() => {
        const slug = repoSlug(owner.value, repo.value);
        return toRef(this.pipelines, slug).value;
      });
    },
    getSortedPipelines(owner: Ref<string>, repo: Ref<string>) {
      return computed(() => Object.values(this.getPipelines(owner, repo).value || []).sort(comparePipelines).sort(comparePipelinesWithStatus));
    },
    getActivePipelines(owner: Ref<string>, repo: Ref<string>) {
      const pipelines = this.getPipelines(owner, repo);
      return computed(() => Object.values(pipelines.value).filter(isPipelineActive));
    },
    getPipeline(owner: Ref<string>, repo: Ref<string>, pipelineNumber: Ref<string>) {
      const pipelines = this.getPipelines(owner, repo);
      return computed(() => (pipelines.value || {})[parseInt(pipelineNumber.value, 10)]);
    },

    // loading
    async loadPipelines(owner: string, repo: string) {
      const pipelines = await apiClient.getPipelineList(owner, repo);
      pipelines.forEach((pipeline) => {
        this.setPipeline(owner, repo, pipeline);
      });
    },
    async loadPipeline(owner: string, repo: string, pipelinesNumber: number) {
      const pipelines = await apiClient.getPipeline(owner, repo, pipelinesNumber);
      this.setPipeline(owner, repo, pipelines);
    },
    async loadPipelineFeed() {
      const pipelines = await apiClient.getPipelineFeed();
      this.pipelineFeed = pipelines;
    },
  },
});
