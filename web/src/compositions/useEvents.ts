import PipelineStore from '~/store/pipelines';
import RepoStore from '~/store/repos';
import { repoSlug } from '~/utils/helpers';

import useApiClient from './useApiClient';

const apiClient = useApiClient();
let initialized = false;

export default () => {
  if (initialized) {
    return;
  }
  const repoStore = RepoStore();
  const pipelineStore = PipelineStore();

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
    pipelineStore.setPipelineFeedItem({ ...pipeline, name: repo.name, owner: repo.owner, full_name: repoSlug(repo) });

    // contains proc update
    if (!data.proc) {
      return;
    }
    const { proc } = data;
    pipelineStore.setProc(repo.owner, repo.name, pipeline.number, proc);
  });
};
