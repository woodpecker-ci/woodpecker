<template>
  <span v-if="started" class="ml-auto text-sm">{{ duration }}</span>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, toRef } from 'vue';

import { useElapsedTime } from '~/compositions/useElapsedTime';
import { PipelineStep, PipelineWorkflow } from '~/lib/api/types';
import { durationAsNumber } from '~/utils/duration';

export default defineComponent({
  name: 'PipelineStepDuration',

  props: {
    step: {
      type: Object as PropType<PipelineStep>,
      default: undefined,
    },

    workflow: {
      type: Object as PropType<PipelineWorkflow>,
      default: undefined,
    },
  },

  setup(props) {
    const step = toRef(props, 'step');
    const workflow = toRef(props, 'workflow');

    const durationRaw = computed(() => {
      const start = (step.value ? step.value?.start_time : workflow.value?.start_time) || 0;
      const end = (step.value ? step.value?.end_time : workflow.value?.end_time) || 0;

      if (end === 0 && start === 0) {
        return undefined;
      }

      if (end === 0) {
        return Date.now() - start * 1000;
      }

      return (end - start) * 1000;
    });

    const running = computed(() => (step.value ? step.value?.state : workflow.value?.state) === 'running');
    const { time: durationElapsed } = useElapsedTime(running, durationRaw);

    const duration = computed(() => {
      if (durationElapsed.value === undefined) {
        return '-';
      }

      return durationAsNumber(durationElapsed.value || 0);
    });
    const started = computed(() => (step.value ? step.value?.start_time : workflow.value?.start_time) !== undefined);

    return { started, duration };
  },
});
</script>
