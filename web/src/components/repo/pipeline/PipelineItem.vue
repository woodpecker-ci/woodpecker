<template>
  <ListItem v-if="pipeline" class="p-0 w-full">
      <div class="flex items-center col-span-4">
        <div
          class="h-full w-2"
          :class="{
            'bg-wp-state-warn-100': pipeline.status === 'pending',
            'bg-wp-state-error-100': pipelineStatusColors[pipeline.status] === 'red',
            'bg-wp-state-neutral-100': pipelineStatusColors[pipeline.status] === 'gray',
            'bg-wp-state-ok-100': pipelineStatusColors[pipeline.status] === 'green',
            'bg-wp-state-info-100': pipelineStatusColors[pipeline.status] === 'blue',
          }"
        />
        <div class="w-12 flex justify-center items-center h-full">
          <PipelineRunningIcon v-if="pipeline.status === 'started' || pipeline.status === 'running'" />
          <PipelineStatusIcon v-else :status="pipeline.status" />
        </div>

        <div class="flex py-2 px-2 mr-2 flex-grow min-w-0 <md:flex-wrap">
          <div class="w-full md:w-auto flex flex-col items-center min-w-0 py-2 gap-2">
            <div class="w-full flex items-center gap-2">
              <span :title="pipelineEventTitle" class="text-wp-text-100">
                <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
                <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
                <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
                <Icon v-else-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />
                <Icon v-else-if="pipeline.event === 'cron'" name="push" />
                <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" />
                <Icon v-else name="push" />
              </span>
              <span class="text-wp-text-100 text-lg whitespace-nowrap overflow-hidden overflow-ellipsis" :title="message">
                {{ shortMessage }}
              </span>
            </div>

            <div class="flex w-full gap-2">
              <div class="flex items-center w-22 text-wp-text-100">
                <Icon name="commit" />
                <span class="truncate">{{ pipeline.commit.slice(0, 7) }}</span>
              </div>

              <div class="flex items-center min-w-0 text-wp-text-100">
                <Icon v-if="pipeline.event === 'pull_request' || pipeline.event === 'pull_request_closed'" name="pull-request" />
                <Icon v-else-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />
                <Icon v-else name="push" />

                <span class="ml-1 truncate">
                  {{ prettyRef }}
                  <!-- eslint-disable @intlify/vue-i18n/no-raw-text-->
                  <span
                    v-if="pipeline.event === 'pull_request' || pipeline.event === 'pull_request_closed'"
                    :title="prTitleWithDescription"
                  >
                    ({{ prTitle }})
                  </span>
                  <!-- eslint-enable @intlify/vue-i18n/no-raw-text-->
                </span>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="<md:hidden flex ml-auto mr-6 items-center">
        <div class="flex flex-col w-full w-48 py-2 gap-x-4 gap-y-2 flex-shrink-0 text-wp-text-100 justify-center h-full">
          <div class="flex space-x-2 min-w-0" :title="$t('repo.pipeline.duration')">
            <Icon name="duration" />
            <span class="truncate">{{ duration }}</span>
          </div>

          <div class="flex space-x-2 min-w-0" :title="$t('repo.pipeline.created', { created })">
            <Icon name="since" />
            <span class="truncate">{{ since }}</span>
          </div>
        </div>

        <Icon v-if="pipeline.event === 'cron'" name="stopwatch" class="text-wp-text-100 w-8" />
        <img v-else class="rounded-md w-8" :src="pipeline.author_avatar" :title="pipeline.author" />
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
const { since, duration, message, shortMessage, prettyRef, created, prTitle, prTitleWithDescription } =
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
