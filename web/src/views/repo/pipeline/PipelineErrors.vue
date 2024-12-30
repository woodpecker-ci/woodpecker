<template>
  <Panel>
    <div class="flex flex-col gap-y-4">
      <template v-for="(error, _index) in pipeline!.errors" :key="_index">
        <div>
          <div class="grid grid-cols-[minmax(10rem,auto),3fr]">
            <span class="flex items-center gap-x-2">
              <Icon
                name="alert"
                class="flex-shrink-0 my-1"
                :class="{
                  'text-wp-state-warn-100': error.is_warning,
                  'text-wp-error-100': !error.is_warning,
                }"
              />
              <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
              <span>
                <code>{{ error.type }}</code>
              </span>
            </span>
            <span
              v-if="isLinterError(error) || isDeprecationError(error) || isBadHabitError(error)"
              class="flex items-center gap-x-2 whitespace-nowrap"
            >
              <span>
                <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
                <span v-if="error.data?.file" class="font-bold">{{ error.data?.file }}: </span>
                <span>{{ error.data?.field }}</span>
              </span>
              <DocsLink
                v-if="isDeprecationError(error) || isBadHabitError(error)"
                :topic="error.data?.field || ''"
                :url="error.data?.docs || ''"
              />
            </span>
            <span v-else />
          </div>
          <div class="col-start-2 grid grid-cols-[minmax(10rem,auto),4fr]">
            <span />
            <span>
              <RenderMarkdown :content="error.message" />
            </span>
          </div>
        </div>
      </template>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { inject, type Ref } from 'vue';

import DocsLink from '~/components/atomic/DocsLink.vue';
import Icon from '~/components/atomic/Icon.vue';
import RenderMarkdown from '~/components/atomic/RenderMarkdown.vue';
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
