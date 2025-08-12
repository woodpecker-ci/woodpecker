<template>
  <aside
    v-if="isOpen"
    ref="target"
    class="border-wp-background-300 dark:border-wp-background-100 bg-wp-background-100 dark:bg-wp-background-300 z-50 flex flex-col items-center overflow-y-auto"
    :aria-label="$t('pipeline_feed')"
  >
    <router-link
      v-for="pipeline in sortedPipelines"
      :key="pipeline.id"
      :to="{
        name: 'repo-pipeline',
        params: { repoId: pipeline.repo_id, pipelineId: pipeline.number },
      }"
      class="border-wp-background-300 dark:border-wp-background-100 hover:bg-wp-control-neutral-100 dark:hover:bg-wp-control-neutral-200 flex w-full border-b px-2 py-4"
    >
      <PipelineFeedItem :pipeline="pipeline" />
    </router-link>

    <span v-if="sortedPipelines.length === 0" class="text-wp-text-100 m-4">{{ $t('repo.pipeline.no_pipelines') }}</span>
  </aside>
</template>

<script lang="ts" setup>
import { onClickOutside } from '@vueuse/core';
import { ref } from 'vue';

import PipelineFeedItem from '~/components/pipeline-feed/PipelineFeedItem.vue';
import usePipelineFeed from '~/compositions/usePipelineFeed';

const pipelineFeed = usePipelineFeed();
const { close, isOpen, sortedPipelines } = pipelineFeed;

const target = ref<HTMLElement>();
onClickOutside(target, close, { ignore: ['.active-pipelines-toggle'] });
</script>
