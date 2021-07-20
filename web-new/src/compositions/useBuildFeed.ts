import { computed, ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Build } from '~/lib/api/types';
import useUserConfig from './useUserConfig';

let initilaized = false;
const builds = ref<Build[] | undefined>();
const activeBuilds = computed(() => {
  if (!builds.value) {
    return undefined;
  }

  return builds.value.filter((build) => ['pending', 'running', 'started'].includes(build.status));
});
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

async function init() {
  const apiClient = useApiClient();

  const b = await apiClient.getBuildFeed();
  builds.value = b.sort(compareFeedItem);

  // listen to build-feed changes
}

export default () => {
  if (!initilaized) {
    init();
  }

  function toggle() {
    setUserConfig('isBuildFeedOpen', !userConfig.value.isBuildFeedOpen);
  }

  return {
    toggle,
    builds,
    activeBuilds,
    isBuildFeedOpen,
  };
};
