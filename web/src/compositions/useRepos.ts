import { useStorage } from '@vueuse/core';
import { ref } from 'vue';

import type { Repo } from '~/lib/api/types';
import { usePipelineStore } from '~/store/pipelines';

export default function useRepos() {
  const pipelineStore = usePipelineStore();
  const lastAccess = useStorage('woodpecker:repo-last-access', new Map<number, number>());

  function repoWithLastPipeline(repo: Repo): Repo {
    if (repo.last_pipeline_number === undefined) {
      return repo;
    }

    if (repo.last_pipeline?.number === repo.last_pipeline_number) {
      return repo;
    }

    const lastPipeline = pipelineStore.getPipeline(ref(repo.id), ref(repo.last_pipeline_number)).value;

    return {
      ...repo,
      last_pipeline: lastPipeline,
    };
  }

  function sortReposByLastAccess(repos: Repo[]): Repo[] {
    return repos
      .filter((r) => lastAccess.value.get(r.id) !== undefined)
      .sort((a, b) => {
        const aLastAccess = lastAccess.value.get(a.id)!;
        const bLastAccess = lastAccess.value.get(b.id)!;

        return bLastAccess - aLastAccess;
      });
  }

  function sortReposByLastActivity(repos: Repo[]): Repo[] {
    return repos.sort((a, b) => {
      const aLastActivity = a.last_pipeline?.created ?? 0;
      const bLastActivity = b.last_pipeline?.created ?? 0;
      return bLastActivity - aLastActivity;
    });
  }

  function updateLastAccess(repoId: number) {
    lastAccess.value.set(repoId, Date.now());
  }

  return {
    sortReposByLastAccess,
    sortReposByLastActivity,
    repoWithLastPipeline,
    updateLastAccess,
  };
}
