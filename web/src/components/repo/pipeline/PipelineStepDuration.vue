<template>
  <span v-if="started" class="ml-auto text-sm">{{ duration }}</span>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import { useDate } from '~/compositions/useDate';
import { useElapsedTime } from '~/compositions/useElapsedTime';
import type { PipelineStep, PipelineWorkflow } from '~/lib/api/types';

const props = defineProps<{
  step?: PipelineStep;
  workflow?: PipelineWorkflow;
}>();

const step = toRef(props, 'step');
const workflow = toRef(props, 'workflow');
const { durationAsNumber } = useDate();

const durationRaw = computed(() => {
  const start = (step.value ? step.value?.started : workflow.value?.started) || 0;
  const end = (step.value ? step.value?.finished : workflow.value?.finished) || 0;

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
const started = computed(() => (step.value ? step.value?.started : workflow.value?.started) !== undefined);
</script>
