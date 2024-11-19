import { defineStore } from 'pinia';
import { computed, reactive, ref, type Ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import type { Repo } from '~/lib/api/types';

import { usePipelineStore } from './pipelines';

export const useRepoStore = defineStore('repos', () => {
  const apiClient = useApiClient();
  const pipelineStore = usePipelineStore();

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
    repos.set(repo.id, {
      ...repos.get(repo.id),
      ...repo,
    });
  }

  async function loadRepo(repoId: number) {
    const repo = await apiClient.getRepo(repoId);
    setRepo(repo);
    return repo;
  }

  async function loadRepos() {
    const _ownedRepos = await apiClient.getRepoList();
    await Promise.all(
      _ownedRepos.map(async (repo) => {
        const lastPipeline = await apiClient.getPipelineList(repo.id, { page: 1, perPage: 1 });
        pipelineStore.setPipeline(repo.id, lastPipeline?.[0]);
        repo.last_pipeline = lastPipeline?.[0].number;
        setRepo(repo);
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
