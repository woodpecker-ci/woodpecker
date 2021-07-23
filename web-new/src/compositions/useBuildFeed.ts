import { computed } from 'vue';
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

export default () => {
  function toggle() {
    setUserConfig('isBuildFeedOpen', !userConfig.value.isBuildFeedOpen);
  }
};
