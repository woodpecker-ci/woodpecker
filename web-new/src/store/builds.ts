import { defineStore } from 'pinia';
import useApiClient from '~/compositions/useApiClient';
import { repoSlug } from '~/compositions/useRepo';
import { Build, BuildProc } from '~/lib/api/types';

const apiClient = useApiClient();

/**
 * Compare two feed items by name.
 * @param {Object} a - A feed item.
 * @param {Object} b - A feed item.
 * @returns {number}
 */
function compareFeedItem(a: Build, b: Build) {
  return (b.started_at || b.created_at || -1) - (a.started_at || a.created_at || -1);
}

export default defineStore({
  id: 'builds',

  state: () => ({
    builds: {} as Record<string, Record<number, Build>>,
  }),

  getters: {
    getActiveBuilds: (state) => (owner: string, repo: string) => {
      return Object.values(state.builds[repoSlug(owner, repo)]).filter((build) =>
        ['pending', 'running', 'started'].includes(build.status),
      );
    },
    getBuild: (state) => (owner: string, repo: string, build: number) => state.builds[repoSlug(owner, repo)][build],
  },

  actions: {
    async setBuild(owner: string, repo: string, build: Build) {
      const _repoSlug = repoSlug(owner, repo);
      if (!this.builds[_repoSlug]) {
        this.builds[_repoSlug] = {};
      }

      this.builds[_repoSlug][build.number] = build;
    },
    async setProc(owner: string, repo: string, build: number, proc: BuildProc) {
      const _repoSlug = repoSlug(owner, repo);
      if (!this.builds[_repoSlug] || !this.builds[_repoSlug][build]) {
        throw new Error("Can't find build");
      }

      const procs = this.builds[_repoSlug][build].procs.filter((p) => p.pid !== proc.pid);
      this.builds[_repoSlug][build].procs = [...procs, proc];
    },
    async loadBuilds(owner: string, repo: string) {
      const b = await apiClient.getBuildList(owner, repo);
      this.builds[repoSlug(owner, repo)] = b.sort(compareFeedItem);
    },
  },
});
