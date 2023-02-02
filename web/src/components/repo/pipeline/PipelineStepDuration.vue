<template>
  <span v-if="step.start_time !== undefined" class="ml-auto text-sm">{{ duration }}</span>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, toRef } from 'vue';

import { useElapsedTime } from '~/compositions/useElapsedTime';
import { PipelineStep } from '~/lib/api/types';
import { durationAsNumber } from '~/utils/duration';

export default defineComponent({
  name: 'PipelineStepDuration',

  props: {
    step: {
      type: Object as PropType<PipelineStep>,
      required: true,
    },
  },

  setup(props) {
    const step = toRef(props, 'step');

    const durationRaw = computed(() => {
      const start = step.value.start_time || 0;
      const end = step.value.end_time || 0;

      if (end === 0 && start === 0) {
        return undefined;
      }

      if (end === 0) {
        return Date.now() - start * 1000;
      }

      return (end - start) * 1000;
    });

    const running = computed(() => step.value.state === 'running');
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
