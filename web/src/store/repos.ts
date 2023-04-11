import { defineStore } from 'pinia';
import { computed, reactive, Ref, ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';

export const useRepoStore = defineStore('repos', () => {
  const apiClient = useApiClient();

  const repos: Map<number, Repo> = reactive(new Map());
  const ownedRepoSlugs = ref<number[]>([]);

  const ownedRepos = computed(() =>
    Array.from(repos.entries())
      .filter(([repoId]) => ownedRepoSlugs.value.includes(repoId))
      .map(([, repo]) => repo),
  );

  function getRepo(repoId: Ref<number>) {
    return computed(() => repos.get(repoId.value));
  }

  function setRepo(repo: Repo) {
    repos.set(repo.id, repo);
  }

  async function loadRepo(repoId: number) {
    const repo = await apiClient.getRepo(repoId);
    repos.set(repo.id, repo);
    return repo;
  }

  async function loadRepos() {
    const _ownedRepos = await apiClient.getRepoList();
    _ownedRepos.forEach((repo) => {
      repos.set(repo.id, repo);
    });
    ownedRepoSlugs.value = _ownedRepos.map((repo) => repo.id);
  }

  return {
    repos,
    ownedRepos,
    ownedRepoSlugs,
    getRepo,
    setRepo,
    loadRepo,
    loadRepos,
  };
});
