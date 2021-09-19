<template>
  <div class="flex mt-4 w-full bg-gray-600 dark:bg-dark-400 min-h-0 flex-grow">
    <div v-if="build.error" class="flex flex-col p-4">
      <span class="text-red-400 font-bold text-xl mb-2">Execution error</span>
      <span class="text-red-400">{{ build.error }}</span>
    </div>

    <div class="flex flex-col w-3/12 text-gray-500">
      <div v-for="proc in build.procs" :key="proc.id">
        <div class="p-4 pb-1">{{ proc.name }}</div>
        <div
          v-for="job in proc.children"
          :key="job.pid"
          class="flex p-2 pl-6 cursor-pointer items-center"
          :class="{ 'bg-gray-800 dark:bg-dark-300': selectedProcId && selectedProcId === job.pid }"
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
          <span v-if="job.start_time !== undefined" class="ml-auto text-gray-500 text-sm">{{ jobDuration(job) }}</span>
        </div>
      </div>
    </div>

    <BuildLogs v-if="selectedProcId" :build="build" :proc-id="selectedProcId" class="w-9/12 flex-grow" />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import BuildLogs from '~/components/repo/build/BuildLogs.vue';
import { Build, BuildProc } from '~/lib/api/types';
import { durationAsNumber } from '~/utils/duration';

export default defineComponent({
  name: 'BuildProcs',

  components: {
    BuildLogs,
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

  setup() {
    function jobDuration(job: BuildProc): string {
      const start = job.start_time || 0;
      const end = job.end_time || 0;

      if (end === 0 && start === 0) {
        return '-';
      }

      if (end === 0) {
        return durationAsNumber(Date.now() - start * 1000);
      }

      return durationAsNumber((end - start) * 1000);
    }

    return { jobDuration };
  },
});
</script>
