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

    <div class="flex flex-grow flex-col relative min-h-0 overflow-y-auto gap-2">
      <div
        v-for="step in pipeline.steps"
        :key="step.id"
        class="p-2 md:rounded-md bg-white shadow dark:border-b-dark-gray-600 dark:bg-dark-gray-700"
      >
        <div class="flex flex-col gap-2">
          <div v-if="step.environ" class="flex flex-wrap gap-x-1 gap-y-2 text-xs justify-end pt-1">
            <div v-for="(value, key) in step.environ" :key="key">
              <span
                class="pl-2 pr-1 py-0.5 bg-gray-800 text-gray-200 dark:bg-gray-600 border-2 border-gray-800 dark:border-gray-600 rounded-l-full"
              >
                {{ key }}
              </span>
              <span class="pl-1 pr-2 py-0.5 border-2 border-gray-800 dark:border-gray-600 rounded-r-full">
                {{ value }}
              </span>
            </div>
          </div>
          <button
            v-if="pipeline.steps && pipeline.steps.length > 1"
            type="button"
            :title="step.name"
            class="flex items-center gap-2 py-2 px-1 hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 rounded-md"
            @click="stepsCollapsed[step.id] = !!!stepsCollapsed[step.id]"
          >
            <Icon
              name="chevron-right"
              class="transition-transform duration-150 min-w-6 h-6"
              :class="{ 'transform rotate-90': !stepsCollapsed[step.id] }"
            />
            <div
              class="min-w-2 h-2 rounded-full -ml-0.75"
              :class="{
                'bg-lime-400': ['success'].includes(step.state),
                'bg-gray-400': ['pending', 'skipped'].includes(step.state),
                'bg-red-400': ['killed', 'error', 'failure', 'blocked', 'declined'].includes(step.state),
                'bg-blue-400': ['started', 'running'].includes(step.state),
              }"
            />
            <span class="truncate">{{ step.name }}</span>
          </button>
        </div>
        <div
          class="transition-height duration-150 overflow-hidden"
          :class="{
            'max-h-screen': !stepsCollapsed[step.id],
            'max-h-0': stepsCollapsed[step.id],
            'ml-6': pipeline.steps && pipeline.steps.length > 1,
          }"
        >
          <button
            v-for="subStep in step.children"
            :key="subStep.pid"
            type="button"
            :title="subStep.name"
            class="flex p-2 gap-2 border-2 border-transparent rounded-md items-center hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 w-full"
            :class="{
              'bg-black bg-opacity-10 dark:bg-white dark:bg-opacity-5':
                selectedStepId && selectedStepId === subStep.pid,
              'mt-1':
                (pipeline.steps && pipeline.steps.length > 1) ||
                (step.children && subStep.pid !== step.children[0].pid),
            }"
            @click="$emit('update:selected-step-id', subStep.pid)"
          >
            <div
              class="min-w-2 h-2 rounded-full"
              :class="{
                'bg-lime-400': ['success'].includes(subStep.state),
                'bg-gray-400': ['pending', 'skipped'].includes(subStep.state),
                'bg-red-400': ['killed', 'error', 'failure', 'blocked', 'declined'].includes(subStep.state),
                'bg-blue-400': ['started', 'running'].includes(subStep.state),
              }"
            />
            <span class="truncate">{{ subStep.name }}</span>
            <PipelineStepDuration :step="subStep" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
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

const stepsCollapsed = ref<Record<PipelineStep['id'], boolean>>(
  props.pipeline.steps && props.pipeline.steps.length > 1
    ? (props.pipeline.steps || []).reduce(
        (collapsed, step) => ({ ...collapsed, [step.id]: !['started', 'running', 'pending'].includes(step.state) }),
        {},
      )
    : {},
);
</script>
