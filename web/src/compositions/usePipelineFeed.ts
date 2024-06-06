import useUserConfig from '~/compositions/useUserConfig';
import { usePipelineStore } from '~/store/pipelines';
import { computed, toRef } from 'vue';
import useAuthentication from './useAuthentication';

const userConfig = useUserConfig();

export default () => {
  const pipelineStore = usePipelineStore();
  const { isAuthenticated } = useAuthentication();

  const isOpen = computed(() => userConfig.userConfig.value.isPipelineFeedOpen && !!isAuthenticated);

  function toggle() {
    userConfig.setUserConfig('isPipelineFeedOpen', !userConfig.userConfig.value.isPipelineFeedOpen);
  }

  function close() {
    userConfig.setUserConfig('isPipelineFeedOpen', false);
  }

  const sortedPipelines = toRef(pipelineStore, 'pipelineFeed');
  const activePipelines = toRef(pipelineStore, 'activePipelines');

  return {
    toggle,
    close,
    isOpen,
    sortedPipelines,
    activePipelines,
    load: pipelineStore.loadPipelineFeed,
  };
};
