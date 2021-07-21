import { defineStore } from 'pinia';
import { computed } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Build } from '~/lib/api/types';
import useUserConfig from '~/compositions/useUserConfig';

const { userConfig, setUserConfig } = useUserConfig();
const isBuildFeedOpen = computed(() => userConfig.value.isBuildFeedOpen);

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
    loaded: false,
    builds: [] as Build[],
  }),

  getters: {
    activeBuilds(state) {
      if (!state.builds) {
        return undefined;
      }

      return state.builds.filter((build) => ['pending', 'running', 'started'].includes(build.status));
    },
    isBuildFeedOpen() {
      return isBuildFeedOpen.value;
    },
  },

  actions: {
    async loadBuilds() {
      if (this.loaded) {
        return;
      }

      this.loaded = true;

      const apiClient = useApiClient();

      const b = await apiClient.getBuildFeed();
      this.builds = b.sort(compareFeedItem);

      // listen to build-feed changes
      apiClient.on((data: any) => {
        console.log('on data', data);
        const { repo, build } = data;
      });
    },
    toggle() {
      setUserConfig('isBuildFeedOpen', !userConfig.value.isBuildFeedOpen);
    },
  },
});
