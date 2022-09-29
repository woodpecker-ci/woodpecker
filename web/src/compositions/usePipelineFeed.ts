import { computed, toRef } from 'vue';

import useUserConfig from '~/compositions/useUserConfig';
import PipelineStore from '~/store/pipelines';

import useAuthentication from './useAuthentication';

const { userConfig, setUserConfig } = useUserConfig();

export default () => {
  const pipelineStore = PipelineStore();
  const { isAuthenticated } = useAuthentication();

  const isOpen = computed(() => userConfig.value.isPipelineFeedOpen && !!isAuthenticated);

  function toggle() {
    setUserConfig('isPipelineFeedOpen', !userConfig.value.isPipelineFeedOpen);
  }

  const sortedPipelines = toRef(pipelineStore, 'sortedPipelineFeed');
  const activePipelines = toRef(pipelineStore, 'activePipelines');

  return {
    toggle,
    isOpen,
    sortedPipelines,
    activePipelines,
    load: pipelineStore.loadPipelineFeed,
  };
};
