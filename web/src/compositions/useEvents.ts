import BuildStore from '~/store/builds';
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
  const buildStore = BuildStore();

  initialized = true;

  apiClient.on((data) => {
    // contains repo update
    if (!data.repo) {
      return;
    }
    const { repo } = data;
    repoStore.setRepo(repo);

    // contains build update
    if (!data.build) {
      return;
    }
    const { build } = data;
    buildStore.setBuild(repo.owner, repo.name, build);
    buildStore.setBuildFeedItem({ ...build, name: repo.name, owner: repo.owner, full_name: repoSlug(repo) });

    // contains proc update
    if (!data.proc) {
      return;
    }
    const { proc } = data;
    buildStore.setProc(repo.owner, repo.name, build.number, proc);
  });
};
