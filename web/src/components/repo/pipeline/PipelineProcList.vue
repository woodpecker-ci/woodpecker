<template>
  <div class="flex flex-col w-full md:w-3/12 md:ml-2 text-gray-600 dark:text-gray-400 gap-2 pb-2">
    <div
      class="flex flex-wrap p-4 gap-1 justify-between flex-shrink-0 md:rounded-md bg-white shadow dark:bg-dark-gray-700"
    >
      <div class="flex space-x-1 items-center flex-shrink-0">
        <div class="flex items-center">
          <Icon v-if="pipeline.event === 'cron'" name="stopwatch" />
          <img v-else class="w-6" :src="pipeline.author_avatar" />
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

    <div v-if="pipeline.procs === undefined || pipeline.procs.length === 0" class="m-auto mt-4">
      <span>{{ $t('repo.pipeline.no_pipeline_steps') }}</span>
    </div>

    <div class="flex flex-grow flex-col relative min-h-0 overflow-y-auto gap-2">
      <div
        v-for="proc in pipeline.procs"
        :key="proc.id"
        class="p-2 md:rounded-md bg-white shadow dark:border-b-dark-gray-600 dark:bg-dark-gray-700"
      >
        <div class="flex flex-col gap-2">
          <div v-if="proc.environ" class="flex flex-wrap gap-x-1 gap-y-2 text-xs justify-end pt-1">
            <div v-for="(value, key) in proc.environ" :key="key">
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
            v-if="pipeline.procs && pipeline.procs.length > 1"
            type="button"
            :title="proc.name"
            class="flex items-center gap-2 py-2 px-1 hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 rounded-md"
            @click="procsCollapsed[proc.id] = !!!procsCollapsed[proc.id]"
          >
            <Icon
              name="chevron-right"
              class="transition-transform duration-150 min-w-6 h-6"
              :class="{ 'transform rotate-90': !procsCollapsed[proc.id] }"
            />
            <span class="truncate">{{ proc.name }}</span>
          </button>
        </div>
        <div
          class="transition-height duration-150 overflow-hidden"
          :class="{
            'max-h-screen': !procsCollapsed[proc.id],
            'max-h-0': procsCollapsed[proc.id],
            'ml-6': pipeline.procs && pipeline.procs.length > 1,
          }"
        >
          <button
            v-for="job in proc.children"
            :key="job.pid"
            type="button"
            :title="job.name"
            class="flex p-2 gap-2 border-2 border-transparent rounded-md items-center hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 w-full"
            :class="{
              'bg-black bg-opacity-10 dark:bg-white dark:bg-opacity-5': selectedProcId && selectedProcId === job.pid,
              'mt-1':
                (pipeline.procs && pipeline.procs.length > 1) || (proc.children && job.pid !== proc.children[0].pid),
            }"
            @click="$emit('update:selected-proc-id', job.pid)"
          >
            <div
              class="min-w-2 h-2 rounded-full"
              :class="{
                'bg-lime-400': ['success'].includes(job.state),
                'bg-gray-400': ['pending', 'skipped'].includes(job.state),
                'bg-red-400': ['killed', 'error', 'failure', 'blocked', 'declined'].includes(job.state),
                'bg-blue-400': ['started', 'running'].includes(job.state),
              }"
            />
            <span class="truncate">{{ job.name }}</span>
            <PipelineProcDuration :proc="job" />
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { ref, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import PipelineProcDuration from '~/components/repo/pipeline/PipelineProcDuration.vue';
import usePipeline from '~/compositions/usePipeline';
import { Pipeline, PipelineProc } from '~/lib/api/types';

const props = defineProps<{
  pipeline: Pipeline;
  selectedProcId?: number | null;
}>();

defineEmits<{
  (event: 'update:selected-proc-id', selectedProcId: number): void;
}>();

const pipeline = toRef(props, 'pipeline');
const { prettyRef } = usePipeline(pipeline);

const procsCollapsed = ref<Record<PipelineProc['id'], boolean>>({});
</script>
