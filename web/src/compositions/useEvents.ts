import { usePipelineStore } from '~/store/pipelines';
import { useRepoStore } from '~/store/repos';

import useApiClient from './useApiClient';

const apiClient = useApiClient();
let initialized = false;

export default () => {
  if (initialized) {
    return;
  }
  const repoStore = useRepoStore();
  const pipelineStore = usePipelineStore();

  initialized = true;

  apiClient.on((data) => {
    // contains repo update
    if (!data.repo) {
      return;
    }
    const { repo } = data;
    repoStore.setRepo(repo);

    // contains pipeline update
    if (!data.pipeline) {
      return;
    }
    const { pipeline } = data;
    pipelineStore.setPipeline(repo.owner, repo.name, pipeline);

    // contains step update
    if (!data.step) {
      return;
    }
    const { step } = data;
    pipelineStore.setWorkflow(repo.owner, repo.name, pipeline.number, step);
  });
};
