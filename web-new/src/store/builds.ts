import { computed, toRef, Ref, ref } from 'vue';
import { defineStore } from 'pinia';
import useApiClient from '~/compositions/useApiClient';
import { repoSlug } from '~/utils/repo';
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

function isBuildActive(build: Build) {
  return ['pending', 'running', 'started'].includes(build.status);
}

export default defineStore({
  id: 'builds',

  state: () => ({
    builds: {} as Record<string, Record<number, Build>>,
    buildFeed: [] as Build[],
  }),

  getters: {
    sortedBuildFeed(state) {
      return state.buildFeed.sort(compareFeedItem);
    },
    activeBuilds(state) {
      return state.buildFeed.filter(isBuildActive);
    },
  },

  actions: {
    // setters
    setBuild(owner: string, repo: string, build: Build) {
      const _repoSlug = repoSlug(owner, repo);
      if (!this.builds[_repoSlug]) {
        this.builds[_repoSlug] = {};
      }

      // const repoBuilds = [...this.builds[_repoSlug].filter((b) => b.id !== build.id), build];
      const repoBuilds = this.builds[_repoSlug];
      repoBuilds[build.number] = build;

      this.builds = {
        ...this.builds,
        [_repoSlug]: repoBuilds,
      };
    },
    setProc(owner: string, repo: string, buildNumber: number, proc: BuildProc) {
      const build = this.getBuild(ref(owner), ref(repo), ref(buildNumber.toString())).value;
      if (!build) {
        throw new Error("Can't find build");
      }

      if (!build.procs) {
        build.procs = [];
      }

      build.procs = [...build.procs.filter((p) => p.pid !== proc.pid), proc];
      this.setBuild(owner, repo, build);
    },

    // getters
    getBuilds(owner: Ref<string>, repo: Ref<string>) {
      return computed(() => {
        const slug = repoSlug(owner.value, repo.value);
        return toRef(this.builds, slug).value;
      });
    },
    getSortedBuilds(owner: Ref<string>, repo: Ref<string>) {
      return computed(() => Object.values(this.getBuilds(owner, repo).value || []).sort(compareFeedItem));
    },
    getActiveBuilds(owner: Ref<string>, repo: Ref<string>) {
      const builds = this.getBuilds(owner, repo);
      return computed(() => Object.values(builds.value).filter(isBuildActive));
    },
    getBuild(owner: Ref<string>, repo: Ref<string>, buildNumber: Ref<string>) {
      const builds = this.getBuilds(owner, repo);
      return computed(() => {
        return (builds.value || {})[parseInt(buildNumber.value)];
      });
    },

    // loading
    async loadBuilds(owner: string, repo: string) {
      const builds = await apiClient.getBuildList(owner, repo);
      builds.forEach((build) => {
        this.setBuild(owner, repo, build);
      });
    },
    async loadBuild(owner: string, repo: string, buildNumber: number) {
      const build = await apiClient.getBuild(owner, repo, buildNumber.toString());
      this.setBuild(owner, repo, build);
    },
    async loadBuildFeed() {
      const builds = await apiClient.getBuildFeed();
      this.buildFeed = builds;
    },
  },
});
