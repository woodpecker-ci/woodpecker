<template>
  <div class="flex mt-4 w-full bg-gray-600 min-h-0 flex-grow">
    <div class="flex flex-col w-3/12 text-white">
      <div v-for="proc in build.procs" :key="proc.id">
        <div class="p-4 pb-1">{{ proc.name }}</div>
        <div
          v-for="job in proc.children"
          :key="job.pid"
          class="flex p-2 pl-6 cursor-pointer items-center"
          :class="{ 'bg-gray-800': selectedProcId && selectedProcId === job.pid }"
          @click="$emit('update:selected-proc-id', job.pid)"
        >
          <div v-if="['success'].includes(job.state)" class="w-2 h-2 bg-status-green rounded-full" />
          <div v-if="['pending', 'skipped'].includes(job.state)" class="w-2 h-2 bg-status-gray rounded-full" />
          <div
            v-if="['killed', 'error', 'failure', 'blocked', 'declined'].includes(job.state)"
            class="w-2 h-2 bg-status-red rounded-full"
          />
          <div v-if="['started', 'running'].includes(job.state)" class="w-2 h-2 bg-status-blue rounded-full" />
          <span class="ml-2">{{ job.name }}</span>
          <span class="ml-auto text-gray-500 text-sm" v-if="job.start_time !== undefined">{{ jobDuration(job) }}</span>
        </div>
      </div>
    </div>

    <BuildLogs v-if="selectedProcId" :build="build" :proc-id="selectedProcId" class="w-9/12 flex-grow" />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import BuildLogs from '~/components/repo/BuildLogs.vue';
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
      required: false,
    },
  },

  emits: {
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
