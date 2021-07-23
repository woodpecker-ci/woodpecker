import { defineStore } from 'pinia';
import useApiClient from '~/compositions/useApiClient';
import { repoSlug } from '~/compositions/useRepo';
import { Repo } from '~/lib/api/types';

const apiClient = useApiClient();

export default defineStore({
  id: 'repos',

  state: () => ({
    repos: {} as Record<string, Repo>,
  }),

  getters: {
    repo: (state) => (owner: string, name: string) => state.repos[repoSlug(owner, name)],
  },

  actions: {
    setRepo(repo: Repo) {
      this.repos[repoSlug(repo)] = repo;
    },
    async loadRepo(owner: string, name: string) {
      const repo = await apiClient.getRepo(owner, name);
      this.repos[repoSlug(repo)] = repo;
    },
  },
});
