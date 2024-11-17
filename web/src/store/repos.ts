import { defineStore } from 'pinia';
import { computed, reactive, ref, type Ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import type { Repo } from '~/lib/api/types';

export const useRepoStore = defineStore('repos', () => {
  const apiClient = useApiClient();

  const repos: Map<number, Repo> = reactive(new Map());
  const ownedRepoIds = ref<number[]>([]);

  const ownedRepos = computed(() =>
    Array.from(repos.entries())
      .filter(([repoId]) => ownedRepoIds.value.includes(repoId))
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
    await Promise.all(
      _ownedRepos.map(async (repo) => {
        const latestPipeline = await apiClient.getPipelineList(repo.id, { page: 1, perPage: 1 });
        repo.last_pipeline_item = latestPipeline[0];
        repos.set(repo.id, repo);
      }),
    );
    ownedRepoIds.value = _ownedRepos.map((repo) => repo.id);
  }

  return {
    repos,
    ownedRepos,
    ownedRepoIds,
    getRepo,
    setRepo,
    loadRepo,
    loadRepos,
  };
});
