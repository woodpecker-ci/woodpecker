<template>
  <div class="md:min-w-xs flex w-full flex-col gap-2 pb-2 text-wp-text-100 md:w-3/12 md:max-w-md">
    <div
      class="flex flex-shrink-0 flex-wrap justify-between gap-1 rounded-md border border-wp-background-400 bg-wp-background-100 p-4 dark:bg-wp-background-200"
    >
      <div class="flex flex-shrink-0 items-center space-x-1">
        <div class="flex items-center">
          <Icon v-if="pipeline.event === 'cron'" name="stopwatch" />
          <img v-else class="w-6 rounded-md" :src="pipeline.author_avatar" />
        </div>
        <span>{{ pipeline.event === 'cron' ? pipeline.cron : pipeline.author }}</span>
      </div>
      <a
        v-if="
          pipeline.event === 'pull_request' ||
          pipeline.event === 'pull_request_closed' ||
          pipeline.event === 'tag' ||
          pipeline.event === 'release'
        "
        class="flex min-w-0 items-center space-x-1 text-wp-link-100 hover:text-wp-link-200"
        :href="pipeline.forge_url"
      >
        <Icon
          v-if="pipeline.event === 'pull_request' || pipeline.event === 'pull_request_closed'"
          name="pull-request"
        />
        <Icon v-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />
        <span class="truncate">{{ prettyRef }}</span>
      </a>
      <router-link
        v-else-if="
          pipeline.event === 'push' ||
          pipeline.event === 'manual' ||
          pipeline.event === 'deployment' ||
          pipeline.event === 'cron'
        "
        class="min-w-0 items-center space-x-1 text-wp-link-100 hover:text-wp-link-200"
        :to="{ name: 'repo-branch', params: { branch: prettyRef } }"
      >
        <Icon v-if="pipeline.event === 'manual'" name="manual-pipeline" />
        <Icon v-else-if="pipeline.event === 'push'" name="push" />
        <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
        <Icon v-else-if="pipeline.event === 'cron'" name="stopwatch" />
        <span class="truncate">{{ prettyRef }}</span>
      </router-link>
      <div class="flex flex-shrink-0 items-center">
        <a
          class="flex items-center text-wp-link-100 hover:text-wp-link-200"
          :href="pipeline.commit.forge_url"
          target="_blank"
        >
          <Icon name="commit" />
          <span>{{ pipeline.commit.sha.slice(0, 10) }}</span>
        </a>
      </div>
    </div>

    <Panel v-if="pipeline.workflows === undefined || pipeline.workflows.length === 0">
      <span>{{ $t('repo.pipeline.no_pipeline_steps') }}</span>
    </Panel>

    <div class="relative min-h-0 w-full flex-grow">
      <div class="absolute left-0 right-0 top-0 flex h-full flex-col gap-y-2 md:overflow-y-auto">
        <div
          v-for="workflow in pipeline.workflows"
          :key="workflow.id"
          class="rounded-md border border-wp-background-400 bg-wp-background-100 p-2 shadow dark:bg-wp-background-200"
        >
          <div class="flex flex-col gap-2">
            <div v-if="workflow.environ" class="flex flex-wrap justify-end gap-x-1 gap-y-2 pr-1 pt-1 text-xs">
              <div v-for="(value, key) in workflow.environ" :key="key">
                <Badge :label="key" :value="value" />
              </div>
            </div>
            <button
              v-if="!singleConfig"
              type="button"
              :title="workflow.name"
              class="hover-effect flex items-center gap-2 rounded-md px-1 py-2 hover:bg-wp-background-300 dark:hover:bg-wp-background-400"
              @click="workflowsCollapsed[workflow.id] = !workflowsCollapsed[workflow.id]"
            >
              <Icon
                name="chevron-right"
                class="h-6 min-w-6 transition-transform duration-150"
                :class="{ 'rotate-90 transform': !workflowsCollapsed[workflow.id] }"
              />
              <PipelineStatusIcon :status="workflow.state" class="!h-4 !w-4" />
              <span class="truncate">{{ workflow.name }}</span>
              <PipelineStepDuration
                v-if="workflow.started !== workflow.finished"
                :workflow="workflow"
                class="pr-2px mr-1"
              />
            </button>
          </div>
          <div
            class="overflow-hidden transition-height duration-150"
            :class="{ 'max-h-0': workflowsCollapsed[workflow.id], 'ml-[1.6rem]': !singleConfig }"
          >
            <button
              v-for="step in workflow.children"
              :key="step.pid"
              type="button"
              :title="step.name"
              class="hover-effect flex w-full items-center gap-2 rounded-md border-2 border-transparent p-2 hover:bg-wp-background-300 dark:hover:bg-wp-background-400"
              :class="{
                'bg-wp-background-300 dark:bg-wp-background-400': selectedStepId && selectedStepId === step.pid,
                'mt-1': !singleConfig || (workflow.children && step.pid !== workflow.children[0].pid),
              }"
              @click="$emit('update:selected-step-id', step.pid)"
            >
              <PipelineStatusIcon :service="step.type === StepType.Service" :status="step.state" class="!h-4 !w-4" />
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
import { computed, inject, ref, toRef } from 'vue';
import type { Ref } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import Panel from '~/components/layout/Panel.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import PipelineStepDuration from '~/components/repo/pipeline/PipelineStepDuration.vue';
import usePipeline from '~/compositions/usePipeline';
import { StepType } from '~/lib/api/types';
import type { Pipeline, PipelineConfig, PipelineStep } from '~/lib/api/types';

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
