<template>
  <div class="flex flex-col w-full md:w-3/12 md:max-w-md md:min-w-xs md:ml-4 text-wp-text-100 gap-2 pb-2">
    <div
      class="flex flex-wrap p-4 gap-1 justify-between flex-shrink-0 rounded-md border bg-wp-background-100 border-wp-background-400 dark:bg-wp-background-200"
    >
      <div class="flex space-x-1 items-center flex-shrink-0">
        <div class="flex items-center">
          <Icon v-if="pipeline.event === 'cron'" name="stopwatch" />
          <img v-else class="rounded-md w-6" :src="pipeline.author_avatar" />
        </div>
        <span>{{ pipeline.author }}</span>
      </div>
      <div>
        <a
          v-if="pipeline.event === 'pull_request'"
          class="flex items-center space-x-1 text-wp-link-100 hover:text-wp-link-200 min-w-0"
          :href="pipeline.link_url"
          target="_blank"
        >
          <Icon name="pull_request" />
          <span class="truncate">{{ prettyRef }}</span>
        </a>
        <router-link
          v-else-if="pipeline.event === 'push' || pipeline.event === 'manual'"
          class="flex items-center space-x-1 text-wp-link-100 hover:text-wp-link-200 min-w-0"
          :to="{ name: 'repo-branch', params: { branch: prettyRef } }"
        >
          <Icon v-if="pipeline.event === 'manual'" name="manual-pipeline" />
          <Icon v-else-if="pipeline.event === 'push'" name="push" />
          <span class="truncate">{{ prettyRef }}</span>
        </router-link>
        <div v-else class="flex space-x-1 items-center min-w-0">
          <Icon v-if="pipeline.event === 'deployment'" name="deployment" />
          <Icon v-else-if="pipeline.event === 'tag'" name="tag" />

          <span class="truncate">{{ prettyRef }}</span>
        </div>
      </div>
      <div class="flex items-center flex-shrink-0">
        <template v-if="pipeline.event === 'pull_request'">
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </template>
        <a
          v-else
          class="text-wp-link-100 hover:text-wp-link-200 flex items-center"
          :href="pipeline.link_url"
          target="_blank"
        >
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </a>
      </div>
    </div>

    <div v-if="pipeline.workflows === undefined || pipeline.workflows.length === 0" class="m-auto mt-4">
      <span>{{ $t('repo.pipeline.no_pipeline_steps') }}</span>
    </div>

    <div class="flex-grow min-h-0 w-full relative">
      <div class="absolute top-0 left-0 right-0 h-full flex flex-col md:overflow-y-auto gap-y-2">
        <div
          v-for="workflow in pipeline.workflows"
          :key="workflow.id"
          class="p-2 rounded-md shadow border bg-wp-background-100 border-wp-background-400 dark:bg-wp-background-200"
        >
          <div class="flex flex-col gap-2">
            <div v-if="workflow.environ" class="flex flex-wrap gap-x-1 gap-y-2 text-xs justify-end pt-1 pr-1">
              <div v-for="(value, key) in workflow.environ" :key="key">
                <Badge :label="key" :value="value" />
              </div>
            </div>
            <button
              v-if="pipeline.workflows && pipeline.workflows.length > 1"
              type="button"
              :title="workflow.name"
              class="flex items-center gap-2 py-2 px-1 hover-effect hover:bg-wp-background-300 dark:hover:bg-wp-background-400 rounded-md"
              @click="workflowsCollapsed[workflow.id] = !workflowsCollapsed[workflow.id]"
            >
              <Icon
                name="chevron-right"
                class="transition-transform duration-150 min-w-6 h-6"
                :class="{ 'transform rotate-90': !workflowsCollapsed[workflow.id] }"
              />
              <PipelineStatusIcon :status="workflow.state" class="!h-4 !w-4" />
              <span class="truncate">{{ workflow.name }}</span>
              <PipelineStepDuration
                v-if="workflow.start_time !== workflow.end_time"
                :workflow="workflow"
                class="mr-1 pr-2px"
              />
            </button>
          </div>
          <div
            class="transition-height duration-150 overflow-hidden"
            :class="{
              'max-h-0': workflowsCollapsed[workflow.id],
              'ml-6': pipeline.workflows && pipeline.workflows.length > 1,
            }"
          >
            <button
              v-for="step in workflow.children"
              :key="step.pid"
              type="button"
              :title="step.name"
              class="flex p-2 gap-2 border-2 border-transparent rounded-md items-center hover-effect hover:bg-wp-background-300 dark:hover:bg-wp-background-400 w-full"
              :class="{
                'bg-wp-background-300 dark:bg-wp-background-400': selectedStepId && selectedStepId === step.pid,
                'mt-1':
                  (pipeline.workflows && pipeline.workflows.length > 1) ||
                  (workflow.children && step.pid !== workflow.children[0].pid),
              }"
              @click="$emit('update:selected-step-id', step.pid)"
            >
              <PipelineStatusIcon :status="step.state" class="!h-4 !w-4" />
              <span class="truncate">{{ step.name }}</span>
              <PipelineStepDuration :step="step" />
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, toRef } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import PipelineStepDuration from '~/components/repo/pipeline/PipelineStepDuration.vue';
import usePipeline from '~/compositions/usePipeline';
import { Pipeline, PipelineStep } from '~/lib/api/types';

const props = defineProps<{
  pipeline: Pipeline;
  selectedStepId?: number | null;
}>();

defineEmits<{
  (event: 'update:selected-step-id', selectedStepId: number): void;
}>();

const pipeline = toRef(props, 'pipeline');
const { prettyRef } = usePipeline(pipeline);

const workflowsCollapsed = ref<Record<PipelineStep['id'], boolean>>(
  props.pipeline.workflows && props.pipeline.workflows.length > 1
    ? (props.pipeline.workflows || []).reduce(
        (collapsed, workflow) => ({
          ...collapsed,
          [workflow.id]: ['success', 'skipped', 'blocked'].includes(workflow.state),
        }),
        {},
      )
    : {},
);
</script>
