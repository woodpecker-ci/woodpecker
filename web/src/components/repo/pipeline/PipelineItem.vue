<template>
  <ListItem v-if="pipeline" class="w-full p-0!">
    <div class="flex w-11 items-center">
      <div
        class="h-full w-3"
        :class="{
          'bg-wp-state-warn-100': pipeline.status === 'pending',
          'bg-wp-error-100 dark:bg-wp-error-200': pipelineStatusColors[pipeline.status] === 'red',
          'bg-wp-state-neutral-100': pipelineStatusColors[pipeline.status] === 'gray',
          'bg-wp-state-ok-100': pipelineStatusColors[pipeline.status] === 'green',
          'bg-wp-state-info-100': pipelineStatusColors[pipeline.status] === 'blue',
        }"
      />
      <div class="flex h-full w-6 flex-wrap items-center justify-between">
        <PipelineRunningIcon v-if="pipeline.status === 'started' || pipeline.status === 'running'" />
        <PipelineStatusIcon v-else class="mx-2 md:mx-3" :status="pipeline.status" />
      </div>
    </div>

    <div class="flex min-w-0 grow flex-wrap px-4 py-2 md:flex-nowrap">
      <div class="hidden shrink-0 items-center md:flex">
        <Icon v-if="pipeline.event === 'cron'" name="stopwatch" class="text-wp-text-100" />
        <img v-else class="w-6 rounded-md" :src="pipeline.author_avatar" />
      </div>

      <div class="flex w-full min-w-0 flex-col gap-y-2 py-2 md:mx-4 md:w-auto">
        <div>
          <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
          <span class="md:display-unset text-wp-text-alt-100 hidden">#{{ pipeline.number }}</span>
          <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
          <span class="md:display-unset text-wp-text-alt-100 mx-2 hidden">-</span>
          <span
            class="text-wp-text-100 overflow-hidden overflow-ellipsis whitespace-nowrap underline md:no-underline"
            :title="message"
          >
            {{ shortMessage }}
          </span>
        </div>

        <div
          v-if="context"
          class="text-wp-text-100 flex items-center gap-x-2 overflow-hidden overflow-ellipsis whitespace-nowrap"
          :title="context"
        >
          <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
          <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
          <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
          <Icon v-else-if="pipeline.event === 'release' || pipeline.event === 'tag'" name="tag" />

          {{ shortContext }}
        </div>
      </div>

      <div
        class="text-wp-text-100 grid w-full shrink-0 grid-flow-col grid-cols-2 grid-rows-2 gap-x-4 gap-y-2 py-2 md:ml-auto md:w-96"
      >
        <div class="flex min-w-0 items-center space-x-2">
          <span :title="pipelineEventTitle">
            <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
            <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
            <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
            <Icon v-else-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />
            <Icon v-else-if="pipeline.event === 'cron'" name="branch" />
            <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" />
            <Icon v-else name="branch" />
          </span>
          <span class="truncate">{{ prettyRef }}</span>
        </div>

        <div class="flex min-w-0 items-center space-x-2">
          <Icon name="commit" />
          <span class="truncate">{{ pipeline.commit.sha.slice(0, 10) }}</span>
        </div>

        <div
          class="flex min-w-0 items-center space-x-2"
          :title="
            durationElapsed > 0 ? $t('repo.pipeline.duration', { duration: durationAsNumber(durationElapsed) }) : ''
          "
        >
          <Icon name="duration" />
          <span class="truncate">{{ duration }}</span>
        </div>

        <div class="flex min-w-0 items-center space-x-2" :title="$t('repo.pipeline.created', { created })">
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
import { useDate } from '~/compositions/useDate';
import usePipeline from '~/compositions/usePipeline';
import type { Pipeline } from '~/lib/api/types';

const props = defineProps<{
  pipeline: Pipeline;
}>();

const { t } = useI18n();
const { durationAsNumber } = useDate();

const pipeline = toRef(props, 'pipeline');
const { since, duration, durationElapsed, message, shortMessage, context, shortContext, prettyRef, created } =
  usePipeline(pipeline);

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
