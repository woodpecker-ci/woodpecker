<template>
  <Panel>
    <div class="grid justify-center gap-2 text-left grid-3-1">
      <template v-for="(error, i) in pipeline.errors" :key="i">
        <span>{{ error.is_warning ? '⚠️' : '❌' }}</span>
        <span>[{{ error.type }}]</span>
        <span v-if="isLinterError(error)" class="underline">
          <span v-if="error.data?.file">{{ error.data?.file }}</span>
          <span>{{ error.data?.field }}</span>
        </span>
        <span v-else />
        <span class="ml-4">{{ error.message }}</span>
      </template>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { inject, Ref } from 'vue';

import Panel from '~/components/layout/Panel.vue';
import type { Pipeline, PipelineError } from '~/lib/api/types';

const pipeline = inject<Ref<Pipeline>>('pipeline');
if (!pipeline) {
  throw new Error('Unexpected: "pipeline" should be provided at this place');
}

type LinterError = PipelineError<{
  file?: string;
  field: string;
}>;

function isLinterError(error: PipelineError): error is LinterError {
  return error.type === 'linter';
}
</script>

<style scoped>
.grid-3-1 {
  grid-template-columns: auto auto auto 1fr;
}
</style>
