<template>
  <div v-if="pipeline" class="flex text-wp-text-100 w-full">
    <PipelineStatusIcon :status="pipeline.status" class="flex items-center" />
    <div class="flex flex-col ml-4 min-w-0">
      <span class="underline">{{ repo?.owner }} / {{ repo?.name }}</span>
      <span class="whitespace-nowrap overflow-hidden overflow-ellipsis">{{ message }}</span>
      <div class="flex flex-col mt-2">
        <div class="flex space-x-2 items-center" :title="created">
          <Icon name="since" />
          <span>{{ since }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <Icon name="duration" />
          <span>{{ duration }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import { PipelineFeed } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';

const props = defineProps<{
  pipeline: PipelineFeed;
}>();

const repoStore = useRepoStore();

const pipeline = toRef(props, 'pipeline');
const repo = repoStore.getRepo(computed(() => pipeline.value.repo_id));

const { since, duration, message, created } = usePipeline(pipeline);
</script>
