import { defineStore } from 'pinia';
import { computed, reactive, ref } from 'vue';
import type { Ref } from 'vue';

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

    _ownedRepos.forEach((repo) => {
      if (repo.last_pipeline) {
        pipelineStore.setPipeline(repo.id, repo.last_pipeline);
        repo.last_pipeline_number = repo.last_pipeline.number;
      }
      setRepo(repo);
    });

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
