<template>
  <div class="flex flex-col gap-2 pb-2 w-full md:w-3/12 md:min-w-xs md:max-w-md text-wp-text-100">
    <div
      class="flex flex-wrap flex-shrink-0 justify-between gap-1 border-wp-background-400 bg-wp-background-200 dark:bg-wp-background-400 p-4 border rounded-md"
    >
      <div class="flex flex-shrink-0 items-center space-x-1">
        <div class="flex items-center">
          <Icon v-if="pipeline.event === 'cron'" name="stopwatch" />
          <img v-else class="rounded-md w-6" :src="pipeline.author_avatar" />
        </div>
        <span>{{ pipeline.author }}</span>
      </div>
      <a
        v-if="pipeline.event === 'pull_request' || pipeline.event === 'pull_request_closed'"
        class="flex items-center space-x-1 min-w-0 text-wp-link-100 hover:text-wp-link-200"
        :href="pipeline.forge_url"
      >
        <Icon name="pull-request" />
        <span class="truncate">{{ prettyRef }}</span>
      </a>
      <router-link
        v-else-if="pipeline.event === 'push' || pipeline.event === 'manual' || pipeline.event === 'deployment'"
        class="flex items-center space-x-1 min-w-0 text-wp-link-100 hover:text-wp-link-200"
        :to="{ name: 'repo-branch', params: { branch: prettyRef } }"
      >
        <Icon v-if="pipeline.event === 'manual'" name="manual-pipeline" />
        <Icon v-else-if="pipeline.event === 'push'" name="push" />
        <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
        <span class="truncate">{{ prettyRef }}</span>
      </router-link>
      <div v-else class="flex items-center space-x-1 min-w-0">
        <Icon v-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />

        <span class="truncate">{{ prettyRef }}</span>
      </div>
      <div class="flex flex-shrink-0 items-center">
        <template v-if="pipeline.event === 'pull_request'">
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </template>
        <a
          v-else
          class="flex items-center text-wp-link-100 hover:text-wp-link-200"
          :href="pipeline.forge_url"
          target="_blank"
        >
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </a>
      </div>
    </div>

    <Panel v-if="pipeline.workflows === undefined || pipeline.workflows.length === 0">
      <span>{{ $t('repo.pipeline.no_pipeline_steps') }}</span>
    </Panel>

    <div class="relative flex-grow w-full min-h-0">
      <div class="top-0 right-0 left-0 absolute flex flex-col gap-y-2 h-full md:overflow-y-auto">
        <div
          v-for="workflow in pipeline.workflows"
          :key="workflow.id"
          class="border-wp-background-400 bg-wp-background-200 dark:bg-wp-background-400 shadow p-2 border rounded-md"
        >
          <div class="flex flex-col gap-2">
            <div v-if="workflow.environ" class="flex flex-wrap justify-end gap-x-1 gap-y-2 pt-1 pr-1 text-xs">
              <div v-for="(value, key) in workflow.environ" :key="key">
                <Badge :label="key" :value="value" />
              </div>
            </div>
            <button
              v-if="!singleConfig"
              type="button"
              :title="workflow.name"
              class="flex items-center gap-2 hover:bg-wp-background-300 dark:hover:bg-wp-background-600 px-1 py-2 rounded-md hover-effect"
              @click="workflowsCollapsed[workflow.id] = !workflowsCollapsed[workflow.id]"
            >
              <Icon
                name="chevron-right"
                class="min-w-6 h-6 transition-transform duration-150"
                :class="{ 'transform rotate-90': !workflowsCollapsed[workflow.id] }"
              />
              <PipelineStatusIcon :status="workflow.state" class="!w-4 !h-4" />
              <span class="truncate">{{ workflow.name }}</span>
              <PipelineStepDuration
                v-if="workflow.started !== workflow.finished"
                :workflow="workflow"
                class="mr-1 pr-2px"
              />
            </button>
          </div>
          <div
            class="transition-height duration-150 overflow-hidden"
            :class="{
              'max-h-0': workflowsCollapsed[workflow.id],
              'ml-6': !singleConfig,
            }"
          >
            <button
              v-for="step in workflow.children"
              :key="step.pid"
              type="button"
              :title="step.name"
              class="flex items-center gap-2 border-2 hover:bg-wp-background-300 dark:hover:bg-wp-background-600 p-2 border-transparent rounded-md w-full hover-effect"
              :class="{
                'bg-wp-background-400 dark:bg-wp-background-500': selectedStepId && selectedStepId === step.pid,
                'mt-1': !singleConfig || (workflow.children && step.pid !== workflow.children[0].pid),
              }"
              @click="$emit('update:selected-step-id', step.pid)"
            >
              <PipelineStatusIcon :service="step.type === StepType.Service" :status="step.state" class="!w-4 !h-4" />
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
import { computed, inject, ref, toRef, type Ref } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import Panel from '~/components/layout/Panel.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import PipelineStepDuration from '~/components/repo/pipeline/PipelineStepDuration.vue';
import usePipeline from '~/compositions/usePipeline';
import { StepType, type Pipeline, type PipelineConfig, type PipelineStep } from '~/lib/api/types';

const props = defineProps<{
  pipeline: Pipeline;
  selectedStepId?: number | null;
}>();

defineEmits<{
  (event: 'update:selected-step-id', selectedStepId: number): void;
}>();

const pipeline = toRef(props, 'pipeline');
const selectedStepId = toRef(props, 'selectedStepId');
const { prettyRef } = usePipeline(pipeline);
const pipelineConfigs = inject<Ref<PipelineConfig[]>>('pipeline-configs');

const workflowsCollapsed = ref<Record<PipelineStep['id'], boolean>>(
  pipeline.value.workflows && pipeline.value.workflows.length > 1
    ? (pipeline.value.workflows || []).reduce(
        (collapsed, workflow) => ({
          ...collapsed,
          [workflow.id]:
            ['success', 'skipped', 'blocked'].includes(workflow.state) &&
            !workflow.children.some((child) => child.pid === selectedStepId.value),
        }),
        {},
      )
    : {},
);

const singleConfig = computed(
  () => pipelineConfigs?.value?.length === 1 && pipeline.value.workflows && pipeline.value.workflows.length === 1,
);
</script>
