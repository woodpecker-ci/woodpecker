<template>
  <div class="flex mt-4 w-full bg-gray-600 dark:bg-dark-gray-800 min-h-0 flex-grow">
    <div v-if="build.error" class="flex flex-col p-4">
      <span class="text-red-400 font-bold text-xl mb-2">Execution error</span>
      <span class="text-red-400">{{ build.error }}</span>
    </div>

    <div class="flex flex-col w-3/12 text-gray-200 dark:text-gray-400">
      <div v-for="proc in build.procs" :key="proc.id">
        <div class="p-4 pb-1 flex flex-wrap items-center justify-between">
          <span>{{ proc.name }}</span>
          <div v-if="proc.environ" class="text-xs">
            <div v-for="(value, key) in proc.environ" :key="key">
              <span
                class="
                  pl-2
                  pr-1
                  py-0.5
                  bg-gray-800
                  dark:bg-gray-600
                  border-2 border-gray-800
                  dark:border-gray-600
                  rounded-l-full
                "
                >{{ key }}</span
              >
              <span class="pl-1 pr-2 py-0.5 border-2 border-gray-800 dark:border-gray-600 rounded-r-full">{{
                value
              }}</span>
            </div>
          </div>
        </div>
        <div
          v-for="job in proc.children"
          :key="job.pid"
          class="flex p-2 pl-6 cursor-pointer items-center hover:bg-gray-700 hover:dark:bg-dark-gray-900"
          :class="{ 'bg-gray-700 !dark:bg-dark-gray-600': selectedProcId && selectedProcId === job.pid }"
          @click="$emit('update:selected-proc-id', job.pid)"
        >
          <div v-if="['success'].includes(job.state)" class="w-2 h-2 bg-lime-400 rounded-full" />
          <div v-if="['pending', 'skipped'].includes(job.state)" class="w-2 h-2 bg-gray-400 rounded-full" />
          <div
            v-if="['killed', 'error', 'failure', 'blocked', 'declined'].includes(job.state)"
            class="w-2 h-2 bg-red-400 rounded-full"
          />
          <div v-if="['started', 'running'].includes(job.state)" class="w-2 h-2 bg-blue-400 rounded-full" />
          <span class="ml-2">{{ job.name }}</span>
          <BuildProcDuration :proc="job" />
        </div>
      </div>
    </div>

    <BuildLogs v-if="selectedProcId" :build="build" :proc-id="selectedProcId" class="w-9/12 flex-grow" />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import BuildLogs from '~/components/repo/build/BuildLogs.vue';
import BuildProcDuration from '~/components/repo/build/BuildProcDuration.vue';
import { Build } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildProcs',

  components: {
    BuildLogs,
    BuildProcDuration,
  },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },

    selectedProcId: {
      type: Number as PropType<number | null>,
      default: null,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:selected-proc-id': (selectedProcId: number) => true,
  },
});
</script>
