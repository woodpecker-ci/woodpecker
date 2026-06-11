import { defineStore } from 'pinia';
import { computed, reactive, ref } from 'vue';
import type { Ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import useConfig from '~/compositions/useConfig';
import { usePaginate } from '~/compositions/usePaginate';
import type { Repo } from '~/lib/api/types';

import { usePipelineStore } from './pipelines';

export const useRepoStore = defineStore('repos', () => {
  const apiClient = useApiClient();
  const pipelineStore = usePipelineStore();

  const repos: Map<number, Repo> = reactive(new Map());
  const ownedRepoIds = ref<number[]>([]);
  const isLoadingRepos = ref(false);
  const isSyncingRepos = ref(false);

  const ownedRepos = computed(() =>
    [...repos.entries()].filter(([repoId]) => ownedRepoIds.value.includes(repoId)).map(([, repo]) => repo),
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
    isLoadingRepos.value = true;
    try {
      return await fetchRepos();
    } finally {
      isLoadingRepos.value = false;
    }
  }

  async function fetchRepos() {
    const _ownedRepos = await apiClient.getRepoList();

    _ownedRepos.forEach((repo) => {
      if (repo.last_pipeline) {
        pipelineStore.setPipeline(repo.id, repo.last_pipeline);
        repo.last_pipeline_number = repo.last_pipeline.number;
      }
      setRepo(repo);
    });

    ownedRepoIds.value = _ownedRepos.map((repo) => repo.id);

    // If the current user is a system admin, also hydrate the store with all repos (paginated)
    const { user } = useConfig();
    const isSystemAdmin = !!user?.admin;
    if (isSystemAdmin) {
      const allRepos = await usePaginate<Repo>(async (page: number) =>
        apiClient.getAllRepos({ page }).then((r) => r ?? []),
      );
      allRepos.forEach((repo) => {
        if (repo.last_pipeline) {
          pipelineStore.setPipeline(repo.id, repo.last_pipeline);
          repo.last_pipeline_number = repo.last_pipeline.number;
        }
        setRepo(repo);
      });
    }

    return _ownedRepos;
  }

  async function loadReposWithBackgroundSync(maxAttempts = 40, pollIntervalMs = 3000) {
    isSyncingRepos.value = false;

    const initialRepos = await loadRepos();
    if (initialRepos.length > 0) {
      return;
    }

    isSyncingRepos.value = true;
    try {
      for (let attempt = 0; attempt < maxAttempts; attempt++) {
        await new Promise((resolve) => {
          setTimeout(resolve, pollIntervalMs);
        });

        const syncedRepos = await fetchRepos();
        if (syncedRepos.length > 0) {
          return;
        }
      }
    } finally {
      isSyncingRepos.value = false;
    }
  }

  return {
    repos,
    ownedRepos,
    ownedRepoIds,
    isLoadingRepos,
    isSyncingRepos,
    getRepo,
    setRepo,
    loadRepo,
    loadRepos,
    loadReposWithBackgroundSync,
  };
});
