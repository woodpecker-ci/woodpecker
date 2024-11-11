<template>
  <Panel>
    <div class="flex flex-col gap-y-4">
      <template v-for="(error,i) in pipeline!.errors" :key="i">
        <div>
          <div class="grid grid-cols-[minmax(10rem,auto),4fr] items-center">
            <span class="flex items-center gap-x-2">
              <Icon
                name="attention"
                class="flex-shrink-0 my-1"
                :class="{
                  'text-wp-state-warn-100': error.is_warning,
                  'text-wp-state-error-100': !error.is_warning,
                }"
              />
              <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
              <span><code>{{ error.type }}</code></span>
            </span>
            <span
              v-if="isLinterError(error) || isDeprecationError(error) || isBadHabitError(error)"
              class="whitespace-nowrap"
            >
              <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
              <span v-if="error.data?.file" class="font-bold">{{ error.data?.file }}: </span>
              <span>{{ error.data?.field }}</span>
            </span>
            <span v-else />
          </div>
          <div class="grid grid-cols-[minmax(10rem,auto),4fr] col-start-2">
            <span />
            <span class="flex gap-x-2">
              <RenderMarkdown :content="error.message" />
              <DocsLink v-if="isDeprecationError(error) || isBadHabitError(error)" :topic="error.data?.field || ''" :url="error.data?.docs || ''" />
            </span>
          </div>
        </div>
      </template>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { inject, type Ref } from 'vue';

import RenderMarkdown from '~/components/atomic/RenderMarkdown.vue';
import DocsLink from '~/components/atomic/DocsLink.vue';
import Icon from '~/components/atomic/Icon.vue';
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

function isBadHabitError(error: PipelineError): error is PipelineError<{ file?: string; field: string; docs: string }> {
  return error.type === 'bad_habit';
}
</script>
