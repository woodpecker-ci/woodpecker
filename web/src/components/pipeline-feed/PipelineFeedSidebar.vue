<template>
  <aside
    v-if="isOpen"
    class="flex flex-col z-50 overflow-y-auto items-center bg-white dark:bg-wp-darkgray-800 dark:border-wp-dark-500"
    :aria-label="$t('pipeline_feed')"
  >
    <router-link
      v-for="pipeline in sortedPipelines"
      :key="pipeline.id"
      :to="{
        name: 'repo-pipeline',
        params: { repoId: pipeline.repo_id, pipelineId: pipeline.number },
      }"
      class="flex border-b py-4 px-2 w-full hover:bg-light-300 dark:hover:bg-wp-darkgray-900 dark:border-wp-darkgray-600 hover:shadow-sm"
    >
      <PipelineFeedItem :pipeline="pipeline" />
    </router-link>

    <span v-if="sortedPipelines.length === 0" class="text-wp-text-100 m-4">{{ $t('repo.pipeline.no_pipelines') }}</span>
  </aside>
</template>

<script lang="ts" setup>
import PipelineFeedItem from '~/components/pipeline-feed/PipelineFeedItem.vue';
import usePipelineFeed from '~/compositions/usePipelineFeed';

const pipelineFeed = usePipelineFeed();
const { isOpen, sortedPipelines } = pipelineFeed;
</script>
