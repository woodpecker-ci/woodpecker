<template>
  <div
    v-if="isPipelineFeedOpen"
    class="flex flex-col z-50 overflow-y-auto items-center bg-white dark:bg-dark-gray-800 dark:border-dark-500"
  >
    <router-link
      v-for="pipeline in sortedPipelineFeed"
      :key="pipeline.id"
      :to="{ name: 'repo-pipeline', params: { repoOwner: pipeline.owner, repoName: pipeline.name, pipelineId: pipeline.number } }"
      class="flex border-b py-4 px-2 w-full hover:bg-light-300 dark:hover:bg-dark-gray-900 dark:border-dark-gray-600 hover:shadow-sm"
    >
      <PipelineFeedItem :pipeline="pipeline" />
    </router-link>

    <span v-if="sortedPipelineFeed.length === 0" class="text-color m-4">{{ $t('repo.pipeline.no_pipelines') }}</span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import PipelineFeedItem from '~/components/pipeline-feed/PipelineFeedItem.vue';
import usePipelineFeed from '~/compositions/usePipelineFeed';

export default defineComponent({
  name: 'PipelineFeedSidebar',

  components: { PipelineFeedItem },

  setup() {
    const pipelineFeed = usePipelineFeed();

    return {
      isPipelineFeedOpen: pipelineFeed.isOpen,
      sortedPipelineFeed: pipelineFeed.sortedPipelines,
    };
  },
});
</script>
