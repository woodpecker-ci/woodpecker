<template>
  <ListItem v-if="pipeline" class="p-0 w-full">
    <div class="flex items-center w-11">
      <div
        class="w-3 h-full"
        :class="{
          'bg-wp-state-warn-100': pipeline.status === 'pending',
          'bg-wp-state-error-100': pipelineStatusColors[pipeline.status] === 'red',
          'bg-wp-state-neutral-100': pipelineStatusColors[pipeline.status] === 'gray',
          'bg-wp-state-ok-100': pipelineStatusColors[pipeline.status] === 'green',
          'bg-wp-state-info-100': pipelineStatusColors[pipeline.status] === 'blue',
        }"
      />
      <div class="flex flex-wrap justify-between items-center w-9 h-full">
        <PipelineRunningIcon v-if="pipeline.status === 'started' || pipeline.status === 'running'" />
        <PipelineStatusIcon v-else class="mx-2 md:mx-3" :status="pipeline.status" />
      </div>
    </div>

    <div class="flex py-2 px-4 flex-grow min-w-0 <md:flex-wrap">
      <div class="<md:hidden flex items-center flex-shrink-0">
        <Icon v-if="pipeline.event === 'cron'" name="stopwatch" class="text-wp-text-100" />
        <img v-else class="rounded-md w-6" :src="pipeline.author_avatar" />
      </div>

      <div class="flex items-center md:mx-4 w-full md:w-auto min-w-0">
        <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
        <span class="text-wp-text-alt-100 <md:hidden">#{{ pipeline.number }}</span>
        <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
        <span class="text-wp-text-alt-100 <md:hidden mx-2">-</span>
        <span
          class="text-wp-text-100 <md:underline whitespace-nowrap overflow-hidden overflow-ellipsis"
          :title="message"
        >
          {{ shortMessage }}
        </span>
      </div>

      <div
        class="flex-shrink-0 gap-x-4 gap-y-2 grid grid-cols-2 grid-rows-2 grid-flow-col md:ml-auto py-2 w-full md:w-96 text-wp-text-100"
      >
        <div class="flex items-center space-x-2 min-w-0">
          <span :title="pipelineEventTitle">
            <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
            <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
            <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
            <Icon v-else-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />
            <Icon v-else-if="pipeline.event === 'cron'" name="push" />
            <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" />
            <Icon v-else name="push" />
          </span>
          <span class="truncate">{{ prettyRef }}</span>
        </div>

        <div class="flex items-center space-x-2 min-w-0">
          <Icon name="commit" />
          <span class="truncate">{{ pipeline.commit.slice(0, 10) }}</span>
        </div>

        <div class="flex items-center space-x-2 min-w-0" :title="$t('repo.pipeline.duration')">
          <Icon name="duration" />
          <span class="truncate">{{ duration }}</span>
        </div>

        <div class="flex items-center space-x-2 min-w-0" :title="$t('repo.pipeline.created', { created })">
          <Icon name="since" />
          <span class="truncate">{{ since }}</span>
        </div>
      </div>
    </div>
  </ListItem>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { pipelineStatusColors } from '~/components/repo/pipeline/pipeline-status';
import PipelineRunningIcon from '~/components/repo/pipeline/PipelineRunningIcon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import type { Pipeline } from '~/lib/api/types';

const props = defineProps<{
  pipeline: Pipeline;
}>();

const { t } = useI18n();

const pipeline = toRef(props, 'pipeline');
const { since, duration, message, shortMessage, prettyRef, created } = usePipeline(pipeline);

const pipelineEventTitle = computed(() => {
  switch (pipeline.value.event) {
    case 'pull_request':
      return t('repo.pipeline.event.pr');
    case 'pull_request_closed':
      return t('repo.pipeline.event.pr_closed');
    case 'deployment':
      return t('repo.pipeline.event.deploy');
    case 'tag':
      return t('repo.pipeline.event.tag');
    case 'release':
      return t('repo.pipeline.event.release');
    case 'cron':
      return t('repo.pipeline.event.cron');
    case 'manual':
      return t('repo.pipeline.event.manual');
    default:
      return t('repo.pipeline.event.push');
  }
});
</script>
