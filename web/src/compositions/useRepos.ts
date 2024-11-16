import { useStorage } from '@vueuse/core';

import type { Repo } from '~/lib/api/types';

export default function useRepos() {
  const lastAccess = useStorage('woodpecker:repo-last-access', new Map<number, number>());

  function sortReposByLastAccess(repos: Repo[]): Repo[] {
    return repos.sort((a, b) => {
      const aLastAccess = lastAccess.value.get(a.id) ?? 0;
      const bLastAccess = lastAccess.value.get(b.id) ?? 0;

      return bLastAccess - aLastAccess;
    });
  }

  function sortReposByLastActivity(repos: Repo[]): Repo[] {
    return repos.sort((a, b) => {
      const aLastActivity = a.last_pipeline || 0;
      const bLastActivity = b.last_pipeline || 0;

      return bLastActivity - aLastActivity;
    });
  }

  function updateLastAccess(repoId: number) {
    lastAccess.value.set(repoId, Date.now());
  }

  return {
    sortReposByLastAccess,
    sortReposByLastActivity,
    updateLastAccess,
  };
}
