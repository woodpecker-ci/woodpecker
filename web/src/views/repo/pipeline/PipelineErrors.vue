<template>
  <Panel>
    <div class="grid justify-center gap-2 text-left grid-3-1">
      <template v-for="(error, i) in pipeline.errors" :key="i">
        <div class="flex items-center gap-4">
          <Icon
            :name="error.is_warning ? 'linter-warn' : 'linter-error'"
            class="flex-shrink-0"
            :class="{
              'text-wp-state-warn-100': error.is_warning,
              'text-wp-state-error-100': !error.is_warning,
            }"
          />
          <span>[{{ error.type }}]</span>
          <span v-if="isLinterError(error) || isDeprecationError(error)">
            <span v-if="error.data?.file" class="font-bold">{{ error.data?.file }}: </span>
            <span>{{ error.data?.field }}</span>
          </span>
          <span v-else />
          <a v-if="isDeprecationError(error)" :href="error.data?.docs" target="_blank" class="underline">
            {{ error.message }}
          </a>
          <span v-else>
            {{ error.message }}
          </span>
        </div>
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

function isLinterError(error: PipelineError): error is PipelineError<{ file?: string; field: string }> {
  return error.type === 'linter';
}

function isDeprecationError(
  error: PipelineError,
): error is PipelineError<{ file: string; field: string; docs: string }> {
  return error.type === 'deprecation';
}
</script>

<style scoped>
.grid-3-1 {
  grid-template-columns: auto auto auto 1fr;
}
</style>
