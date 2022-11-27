import { defineStore } from 'pinia';
import { computed, reactive, Ref, ref } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import { repoSlug } from '~/utils/helpers';

export const useRepoStore = defineStore('repos', () => {
  const apiClient = useApiClient();

  const repos: Map<string, Repo> = reactive(new Map());
  const ownedRepoSlugs = ref<string[]>([]);

  const ownedRepos = computed(() =>
    Array.from(repos.entries())
      .filter(([slug]) => ownedRepoSlugs.value.includes(slug))
      .map(([, repo]) => repo),
  );

  function getRepo(owner: Ref<string>, name: Ref<string>) {
    return computed(() => {
      const slug = repoSlug(owner.value, name.value);
      return repos.get(slug);
    });
  }

  function setRepo(repo: Repo) {
    repos.set(repoSlug(repo), repo);
  }

  async function loadRepo(owner: string, name: string) {
    const repo = await apiClient.getRepo(owner, name);
    repos.set(repoSlug(repo), repo);
    return repo;
  }

  async function loadRepos() {
    const _repos = await apiClient.getRepoList();
    _repos.forEach((repo) => {
      repos.set(repoSlug(repo), repo);
    });
    ownedRepoSlugs.value = _repos.map((repo) => repoSlug(repo));
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
