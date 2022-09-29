<template>
  <span v-if="proc.start_time !== undefined" class="ml-auto text-sm">{{ duration }}</span>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, toRef } from 'vue';

import { useElapsedTime } from '~/compositions/useElapsedTime';
import { PipelineProc } from '~/lib/api/types';
import { durationAsNumber } from '~/utils/duration';

export default defineComponent({
  name: 'PipelineProcDuration',

  props: {
    proc: {
      type: Object as PropType<PipelineProc>,
      required: true,
    },
  },

  setup(props) {
    const proc = toRef(props, 'proc');

    const durationRaw = computed(() => {
      const start = proc.value.start_time || 0;
      const end = proc.value.end_time || 0;

      if (end === 0 && start === 0) {
        return undefined;
      }

      if (end === 0) {
        return Date.now() - start * 1000;
      }

      return (end - start) * 1000;
    });

    const running = computed(() => proc.value.state === 'running');
    const { time: durationElapsed } = useElapsedTime(running, durationRaw);

    const duration = computed(() => {
      if (durationElapsed.value === undefined) {
        return '-';
      }

      return durationAsNumber(durationElapsed.value);
    });

    return { duration };
  },
});
</script>
