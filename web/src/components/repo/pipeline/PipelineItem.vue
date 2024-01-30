<template>
  <ListItem v-if="pipeline" class="p-0 w-full">
    <div class="flex w-6 items-center md:mr-4">
      <div
        class="h-full w-2 flex-shrink-0"
        :class="{
          'bg-wp-state-warn-100': pipeline.status === 'pending',
          'bg-wp-state-error-100': pipelineStatusColors[pipeline.status] === 'red',
          'bg-wp-state-neutral-100': pipelineStatusColors[pipeline.status] === 'gray',
          'bg-wp-state-ok-100': pipelineStatusColors[pipeline.status] === 'green',
          'bg-wp-state-info-100': pipelineStatusColors[pipeline.status] === 'blue',
        }"
      />
      <div class="w-4 flex flex-wrap justify-between items-center h-full">
        <PipelineRunningIcon v-if="pipeline.status === 'started' || pipeline.status === 'running'" />
        <PipelineStatusIcon v-else class="mx-2 md:mx-3" :status="pipeline.status" />
      </div>
    </div>

    <div class="flex py-2 px-4 flex-grow min-w-0 <md:flex-wrap">
      <div class="flex flex-col justify-center w-4 flex-shrink-0">
        <span class="<md:hidden">#{{ pipeline.number }}</span>
        <span class="text-wp-text-100 truncate">{{ duration }}</span>
      </div>

      <div class="flex flex-col min-w-0 justify-center">
        <span
          class="text-wp-text-100 <md:underline whitespace-nowrap overflow-hidden overflow-ellipsis"
          :title="message"
          >{{ title }}
        </span>
      </div>

      <div class="flex <md:flex-col ml-auto w-3/10 min-w-0 flex-shrink-0 gap-2 justify-between">
        <div class="flex gap-2 items-center">
          <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
          <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
          <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
          <Icon v-else-if="pipeline.event === 'tag'" name="tag" />
          <Icon v-else-if="pipeline.event === 'cron'" name="push" />
          <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" />
          <Icon v-else name="push" />
          <span class="truncate">{{ prettyRef }}</span>
        </div>

        <div class="flex items-center gap-1 min-w-0">
          <span class="text-wp-text-100 truncate overflow-ellipsis gap-1 flex">{{ since }} by {{ pipeline.author }}</span>

          <Icon v-if="pipeline.event === 'cron'" name="stopwatch" class="text-wp-text-100" />
          <img v-else class="rounded-md w-5" :src="pipeline.author_avatar" :title="pipeline.author" />
        </div>
      </div>

      <!-- <div
        class="grid grid-cols-2 grid-flow-row w-full md:ml-auto md:w-96 py-2 gap-x-4 gap-y-2 flex-shrink-0 text-wp-text-100"
      >
        <div class="flex gap-2 items-center min-w-0">
          <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
          <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
          <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
          <Icon v-else-if="pipeline.event === 'tag'" name="tag" />
          <Icon v-else-if="pipeline.event === 'cron'" name="push" />
          <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" />
          <Icon v-else name="push" />
          <span class="truncate">{{ prettyRef }}</span>
        </div>

        <div class="flex gap-2 items-center min-w-0" :title="$t('commit_sha')">
          <Icon name="commit" />
          <span class="truncate">{{ pipeline.commit.slice(0, 10) }}</span>
        </div>

        <div class="flex gap-2 items-center min-w-0" :title="$t('pipeline_duration')">
          <Icon name="duration" />
          <span class="truncate">{{ duration }}</span>
        </div>

        <div class="flex gap-2 items-center min-w-0" :title="$t('pipeline_since', { created })">
          <Icon name="since" />
          <span>{{ since }}</span>
        </div>
      </div> -->
    </div>
  </ListItem>
</template>

<script lang="ts" setup>
import { toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { pipelineStatusColors } from '~/components/repo/pipeline/pipeline-status';
import PipelineRunningIcon from '~/components/repo/pipeline/PipelineRunningIcon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import { Pipeline } from '~/lib/api/types';

const props = defineProps<{
  pipeline: Pipeline;
}>();

const pipeline = toRef(props, 'pipeline');
const { since, duration, message, title, prettyRef, created } = usePipeline(pipeline);
</script>
