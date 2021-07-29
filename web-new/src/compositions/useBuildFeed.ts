import { computed, toRef } from 'vue';
import BuildStore from '~/store/builds';
import useUserConfig from '~/compositions/useUserConfig';

const { userConfig, setUserConfig } = useUserConfig();
const isOpen = computed(() => userConfig.value.isBuildFeedOpen);

export default () => {
  const buildStore = BuildStore();

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
    load: buildStore.loadBuildFeed,
  };
};
