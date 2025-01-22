import { useStorage } from '@vueuse/core';
import { ref } from 'vue';

import type { Repo } from '~/lib/api/types';
import { usePipelineStore } from '~/store/pipelines';

export default function useRepos() {
  const pipelineStore = usePipelineStore();
  const lastAccess = useStorage('woodpecker:repo-last-access', new Map<number, number>());

  function repoWithLastPipeline(repo: Repo): Repo {
    if (repo.last_pipeline === undefined) {
      return repo;
    }

    if (repo.last_pipeline_item?.number === repo.last_pipeline) {
      return repo;
    }

    const lastPipeline = pipelineStore.getPipeline(ref(repo.id), ref(repo.last_pipeline)).value;

    return {
      ...repo,
      last_pipeline_item: lastPipeline,
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
      const aLastActivity = a.last_pipeline_item?.created ?? 0;
      const bLastActivity = b.last_pipeline_item?.created ?? 0;
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
