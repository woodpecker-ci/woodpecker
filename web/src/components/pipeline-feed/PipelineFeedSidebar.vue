<template>
  <div
    v-if="isBuildFeedOpen"
    class="flex flex-col z-50 overflow-y-auto items-center bg-white dark:bg-dark-gray-800 dark:border-dark-500"
  >
    <router-link
      v-for="build in sortedBuildFeed"
      :key="build.id"
      :to="{ name: 'repo-pipeline', params: { repoOwner: build.owner, repoName: build.name, pipelineId: build.number } }"
      class="flex border-b py-4 px-2 w-full hover:bg-light-300 dark:hover:bg-dark-gray-900 dark:border-dark-gray-600 hover:shadow-sm"
    >
      <BuildFeedItem :build="build" />
    </router-link>

    <span v-if="sortedBuildFeed.length === 0" class="text-color m-4">{{ $t('repo.pipeline.no_pipelines') }}</span>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import BuildFeedItem from '~/components/pipeline-feed/PipelineFeedItem.vue';
import useBuildFeed from '~/compositions/usePipelineFeed';

export default defineComponent({
  name: 'BuildFeedSidebar',

  components: { BuildFeedItem },

  setup() {
    const buildFeed = useBuildFeed();

    return {
      isBuildFeedOpen: buildFeed.isOpen,
      sortedBuildFeed: buildFeed.sortedBuilds,
    };
  },
});
</script>
