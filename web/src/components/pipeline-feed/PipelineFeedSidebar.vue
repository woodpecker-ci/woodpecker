<template>
  <aside
    v-if="isOpen"
    v-on-click-outside="close"
    class="flex flex-col z-50 overflow-y-auto items-center bg-wp-background-100 dark:bg-wp-background-200 border-wp-background-400"
    :aria-label="$t('pipeline_feed')"
  >
    <router-link
      v-for="pipeline in sortedPipelines"
      :key="pipeline.id"
      :to="{
        name: 'repo-pipeline',
        params: { repoId: pipeline.repo_id, pipelineId: pipeline.number },
      }"
      class="flex border-b border-wp-background-400 py-4 px-2 w-full hover:bg-wp-background-300 dark:hover:bg-wp-background-400 hover:shadow-sm"
    >
      <PipelineFeedItem :pipeline="pipeline" />
    </router-link>

    <span v-if="sortedPipelines.length === 0" class="text-wp-text-100 m-4">{{ $t('repo.pipeline.no_pipelines') }}</span>
  </aside>
</template>

<script lang="ts" setup>
import { vOnClickOutside } from '@vueuse/components';

import PipelineFeedItem from '~/components/pipeline-feed/PipelineFeedItem.vue';
import usePipelineFeed from '~/compositions/usePipelineFeed';

const pipelineFeed = usePipelineFeed();
const { close, isOpen, sortedPipelines } = pipelineFeed;
</script>
