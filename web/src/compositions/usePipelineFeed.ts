import { computed, toRef } from 'vue';

import useUserConfig from '~/compositions/useUserConfig';
import BuildStore from '~/store/pipelines';

import useAuthentication from './useAuthentication';

const { userConfig, setUserConfig } = useUserConfig();

export default () => {
  const buildStore = BuildStore();
  const { isAuthenticated } = useAuthentication();

  const isOpen = computed(() => userConfig.value.isBuildFeedOpen && !!isAuthenticated);

  function toggle() {
    setUserConfig('isBuildFeedOpen', !userConfig.value.isBuildFeedOpen);
  }

  const sortedBuilds = toRef(buildStore, 'sortedBuildFeed');
  const activeBuilds = toRef(buildStore, 'activeBuilds');

  return {
    toggle,
    isOpen,
    sortedBuilds,
    activeBuilds,
    load: buildStore.loadPipelineFeed,
  };
};
