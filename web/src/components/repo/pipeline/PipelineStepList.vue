<template>
  <div class="flex flex-col w-full md:w-3/12 md:ml-2 text-gray-600 dark:text-gray-400 gap-2 pb-2">
    <div
      class="flex flex-wrap p-4 gap-1 justify-between flex-shrink-0 md:rounded-md bg-white shadow dark:bg-dark-gray-700"
    >
      <div class="flex space-x-1 items-center flex-shrink-0">
        <div class="flex items-center">
          <Icon v-if="pipeline.event === 'cron'" name="stopwatch" />
          <img v-else class="rounded-md w-6" :src="pipeline.author_avatar" />
        </div>
        <span>{{ pipeline.author }}</span>
      </div>
      <div class="flex space-x-1 items-center min-w-0">
        <Icon v-if="pipeline.event === 'manual'" name="manual-pipeline" />
        <Icon v-if="pipeline.event === 'push'" name="push" />
        <Icon v-if="pipeline.event === 'deployment'" name="deployment" />
        <Icon v-else-if="pipeline.event === 'tag'" name="tag" />
        <a
          v-else-if="pipeline.event === 'pull_request'"
          class="flex items-center space-x-1 text-link min-w-0"
          :href="pipeline.link_url"
          target="_blank"
        >
          <Icon name="pull_request" />
          <span class="truncate">{{ prettyRef }}</span>
        </a>
        <span v-if="pipeline.event !== 'pull_request'" class="truncate">{{ pipeline.branch }}</span>
      </div>
      <div class="flex items-center flex-shrink-0">
        <template v-if="pipeline.event === 'pull_request'">
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </template>
        <a v-else class="text-blue-700 dark:text-link flex items-center" :href="pipeline.link_url" target="_blank">
          <Icon name="commit" />
          <span>{{ pipeline.commit.slice(0, 10) }}</span>
        </a>
      </div>
    </div>

    <div v-if="pipeline.steps === undefined || pipeline.steps.length === 0" class="m-auto mt-4">
      <span>{{ $t('repo.pipeline.no_pipeline_steps') }}</span>
    </div>

    <div class="flex-grow min-h-0 w-full relative">
      <div class="absolute top-0 left-0 right-0 h-full flex flex-col overflow-y-scroll gap-y-2">
        <div
          v-for="workflow in pipeline.steps"
          :key="workflow.id"
          class="p-2 md:rounded-md bg-white shadow dark:border-b-dark-gray-600 dark:bg-dark-gray-700"
        >
          <div class="flex flex-col gap-2">
            <div v-if="workflow.environ" class="flex flex-wrap gap-x-1 gap-y-2 text-xs justify-end pt-1 pr-1">
              <div v-for="(value, key) in workflow.environ" :key="key">
                <Badge :label="key" :value="value" />
              </div>
            </div>
            <button
              v-if="pipeline.steps && pipeline.steps.length > 1"
              type="button"
              :title="workflow.name"
              class="flex items-center gap-2 py-2 px-1 hover-effect rounded-md"
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
                :step="workflow"
                class="mr-1 pr-2px"
              />
              <button
                v-if="pipeline.steps && pipeline.steps.length > 0 && ['pending'].includes(workflow.state)"
                :title="$t('repo.pipeline.actions.skip')"
                type="button"
                class="flex justify-center items-center ml-auto gap-2 py-2 px-1 hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 rounded-md"
                @click="skipWorkflow(workflow)"
              >
                <Icon name="status-skipped" class="!h-4 !w-4" />
              </button>
            </button>
          </div>
          <div
            class="transition-height duration-150 overflow-hidden"
            :class="{
              'max-h-screen': !workflowsCollapsed[workflow.id],
              'max-h-0': workflowsCollapsed[workflow.id],
              'ml-6': pipeline.steps && pipeline.steps.length > 1,
            }"
          >
            <button
              v-for="step in workflow.children"
              :key="step.pid"
              type="button"
              :title="step.name"
              class="flex p-2 gap-2 border-2 border-transparent rounded-md items-center hover-effect w-full"
              :class="{
                'bg-black bg-opacity-10 dark:bg-white dark:bg-opacity-5': selectedStepId && selectedStepId === step.pid,
                'mt-1':
                  (pipeline.steps && pipeline.steps.length > 1) ||
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
import { inject, Ref, ref, toRef } from 'vue';

import { useI18n } from 'vue-i18n';
import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import PipelineStepDuration from '~/components/repo/pipeline/PipelineStepDuration.vue';
import usePipeline from '~/compositions/usePipeline';
import useApiClient from '~/compositions/useApiClient';
import { Pipeline, PipelineStep, Repo } from '~/lib/api/types';
import useNotifications from '~/compositions/useNotifications';
import Button from "~/components/atomic/Button.vue";

const props = defineProps<{
  pipeline: Pipeline;
  selectedStepId?: number | null;
}>();

defineEmits<{
  (event: 'update:selected-step-id', selectedStepId: number): void;
}>();

const i18n = useI18n();
const notifications = useNotifications();
const apiClient = useApiClient();
const pipeline = toRef(props, 'pipeline');
const { prettyRef } = usePipeline(pipeline);
const repo = inject<Ref<Repo>>('repo');

if (!repo) {
  throw new Error('Unexpected: "repo", "repoPermissions" & "pipeline" should be provided at this place');
}

const workflowsCollapsed = ref<Record<PipelineStep['id'], boolean>>(
  props.pipeline.steps && props.pipeline.steps.length > 1
    ? (props.pipeline.steps || []).reduce(
        (collapsed, workflow) => ({
          ...collapsed,
          [workflow.id]: ['success', 'skipped', 'blocked'].includes(workflow.state),
        }),
        {},
      )
    : {},
);

const skipWorkflow = async (workflow: PipelineStep) => {
  await apiClient.skipPipelineWorkflow(repo.value.owner, repo.value.name, `${pipeline.value.number}`, workflow.pid);
  notifications.notify({ title: i18n.t('repo.pipeline.actions.skip_success'), type: 'success' });
}

</script>
