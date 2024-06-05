import { computed, toRef } from 'vue';
import useAuthentication from './useAuthentication';
import useUserConfig from '~/compositions/useUserConfig';
import { usePipelineStore } from '~/store/pipelines';

const { userConfig, setUserConfig } = useUserConfig();

export default () => {
  const pipelineStore = usePipelineStore();
  const { isAuthenticated } = useAuthentication();

  const isOpen = computed(() => userConfig.value.isPipelineFeedOpen && !!isAuthenticated);

  function toggle() {
    setUserConfig('isPipelineFeedOpen', !userConfig.value.isPipelineFeedOpen);
  }

  function close() {
    setUserConfig('isPipelineFeedOpen', false);
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
