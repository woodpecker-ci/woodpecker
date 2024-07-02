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
      <div class="w-6 flex flex-wrap justify-between items-center h-full">
        <PipelineRunningIcon v-if="pipeline.status === 'started' || pipeline.status === 'running'" />
        <PipelineStatusIcon v-else class="mx-2 md:mx-3" :status="pipeline.status" />
      </div>
    </div>

    <div class="flex py-2 px-4 flex-grow min-w-0 <md:flex-wrap">
      <div class="flex flex-col min-w-0 justify-center gap-2">
        <span
          class="text-wp-text-100 text-lg <md:underline whitespace-nowrap overflow-hidden overflow-ellipsis"
          :title="message"
        >
          {{ shortMessage }}
        </span>
        <div class="flex gap-1 text-wp-text-alt-100">
          <div class="flex items-center" :title="pipelineEventTitle">
            <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
            <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
            <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
            <Icon v-else-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />
            <Icon v-else-if="pipeline.event === 'cron'" name="push" />
            <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" />
            <Icon v-else name="push" />
          </div>

          <a :href="pipeline.forge_url" target="_blank" class="underline" :title="pipeline.commit">
            <Badge :label="pipeline.commit.slice(0, 7)" />
          </a>

          <span v-if="pipeline.event === 'pull_request' || pipeline.event === 'push'">{{ $t('pushed_to') }}</span>
          <span v-if="pipeline.event === 'pull_request_closed'">{{ $t('closed') }}</span>
          <span v-if="pipeline.event === 'deployment'">{{ $t('deployed_to') }}</span>
          <span v-if="pipeline.event === 'tag' || pipeline.event === 'release'">{{ $t('created') }}</span>
          <span v-if="pipeline.event === 'cron' || pipeline.event === 'manual'">{{ $t('triggered') }}</span>
          <span v-else>{{ $t('triggered') }}</span>
          <Badge v-if="prettyRef" :title="title" :label="title ? `${prettyRef} (${truncate(title, 30)})` : prettyRef " />
          <span class="truncate">{{ $t('by_user', { user: pipeline.author }) }}</span>
        </div>
      </div>

      <div class="flex <md:flex-col min-w-0 gap-2 justify-between items-center ml-auto">
        <div class="flex flex-col gap-2 text-wp-text-alt-100 mr-4">
          <div class="flex gap-2 items-center min-w-0" :title="$t('pipeline_duration')">
            <Icon name="duration" />
            <span class="truncate">{{ duration }}</span>
          </div>

          <div class="flex gap-2 items-center min-w-0" :title="$t('pipeline_since', { created })">
            <Icon name="since" />
            <span>{{ since }}</span>
          </div>
        </div>

        <Icon v-if="pipeline.event === 'cron'" name="stopwatch" class="text-wp-text-100" />
        <img v-else class="rounded-md w-8 flex-shrink-0" :src="pipeline.author_avatar" :title="pipeline.author" />
      </div>
    </div>
  </ListItem>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { pipelineStatusColors } from '~/components/repo/pipeline/pipeline-status';
import PipelineRunningIcon from '~/components/repo/pipeline/PipelineRunningIcon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import type { Pipeline } from '~/lib/api/types';
import { truncate } from '~/utils/locale';

const props = defineProps<{
  pipeline: Pipeline;
}>();

const pipeline = toRef(props, 'pipeline');
const { since, duration, message, shortMessage, title, prettyRef, created } = usePipeline(pipeline);

const pipelineEventTitle = computed(() => {
  switch (pipeline.value.event) {
    case 'pull_request':
      return 'Pull request'; // TODO: translate
    case 'pull_request_closed':
      return 'Pull request closed / merged';
    case 'deployment':
      return 'Deployment';
    case 'tag':
    case 'release':
      return 'Tag';
    case 'cron':
      return 'Cron';
    case 'manual':
      return 'Manual';
    default:
      return 'Push';
  }
});
</script>
