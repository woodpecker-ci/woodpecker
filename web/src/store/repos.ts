import { defineStore } from 'pinia';
import { computed, Ref, toRef } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import { repoSlug } from '~/utils/helpers';

const apiClient = useApiClient();

export default defineStore({
  id: 'repos',

  state: () => ({
    repos: {} as Record<string, Repo>,
  }),

  actions: {
    // getter
    getRepo(owner: Ref<string>, name: Ref<string>) {
      return computed(() => {
        const slug = repoSlug(owner.value, name.value);
        return toRef(this.repos, slug).value;
      });
    },

    // setter
    setRepo(repo: Repo) {
      this.repos[repoSlug(repo)] = repo;
    },

    // loading
    async loadRepo(owner: string, name: string) {
      const repo = await apiClient.getRepo(owner, name);
      this.repos[repoSlug(repo)] = repo;
      return repo.full_name;
    },
    async loadRepos() {
      const repos = await apiClient.getRepoList();
      repos.forEach((repo) => {
        this.repos[repoSlug(repo.owner, repo.name)] = repo;
      });
    },
  },
});
