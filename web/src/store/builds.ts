import { defineStore } from 'pinia';
import { computed, Ref, ref, toRef } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Build, BuildFeed, BuildProc } from '~/lib/api/types';
import { compareBuilds, isBuildActive, repoSlug } from '~/utils/helpers';

const apiClient = useApiClient();

export default defineStore({
  id: 'builds',

  state: () => ({
    builds: {} as Record<string, Record<number, Build>>,
    buildFeed: [] as BuildFeed[],
  }),

  getters: {
    sortedBuildFeed(state) {
      return state.buildFeed.sort(compareBuilds);
    },
    activeBuilds(state) {
      return state.buildFeed.filter(isBuildActive);
    },
  },

  actions: {
    // setters
    setBuild(owner: string, repo: string, build: Build) {
      // eslint-disable-next-line @typescript-eslint/naming-convention
      const _repoSlug = repoSlug(owner, repo);
      if (!this.builds[_repoSlug]) {
        this.builds[_repoSlug] = {};
      }

      const repoBuilds = this.builds[_repoSlug];

      // merge with available data
      repoBuilds[build.number] = { ...(repoBuilds[build.number] || {}), ...build };

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
    setBuildFeedItem(build: BuildFeed) {
      const buildFeed = this.buildFeed.filter((b) => b.id !== build.id);
      this.buildFeed = [...buildFeed, build];
    },

    // getters
    getBuilds(owner: Ref<string>, repo: Ref<string>) {
      return computed(() => {
        const slug = repoSlug(owner.value, repo.value);
        return toRef(this.builds, slug).value;
      });
    },
    getSortedBuilds(owner: Ref<string>, repo: Ref<string>) {
      return computed(() => Object.values(this.getBuilds(owner, repo).value || []).sort(compareBuilds));
    },
    getActiveBuilds(owner: Ref<string>, repo: Ref<string>) {
      const builds = this.getBuilds(owner, repo);
      return computed(() => Object.values(builds.value).filter(isBuildActive));
    },
    getBuild(owner: Ref<string>, repo: Ref<string>, buildNumber: Ref<string>) {
      const builds = this.getBuilds(owner, repo);
      return computed(() => (builds.value || {})[parseInt(buildNumber.value, 10)]);
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
