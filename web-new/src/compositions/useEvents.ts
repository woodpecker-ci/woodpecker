import useApiClient from './useApiClient';
import RepoStore from '~/store/repos';
import BuildStore from '~/store/builds';

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

    // contains proc update
    if (!data.proc) {
      return;
    }
    const { proc } = data;
    buildStore.setProc(repo.owner, repo.name, build.number, proc);
  });
};
