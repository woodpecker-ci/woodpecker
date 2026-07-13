<template>
  <Panel>
    <div class="flex flex-col gap-y-4">
      <!-- Hard parse errors block the pipeline and come first; if parsing only
           produced warnings, the runtime errors are the real failure cause and
           are shown on top instead. -->
      <div
        v-if="pipeline!.errors && pipeline!.errors.length > 0"
        class="flex flex-col gap-y-4"
        :class="hasHardParseErrors(pipeline) ? 'order-1' : 'order-2'"
      >
        <span class="text-lg font-bold">{{ $t('repo.pipeline.parse_errors') }}</span>
        <template v-for="(error, _index) in pipeline!.errors" :key="_index">
          <div>
            <div class="grid grid-cols-[minmax(10rem,auto)_3fr]">
              <span class="flex items-center gap-x-2">
                <Icon
                  name="alert"
                  class="my-1 shrink-0"
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
            <div class="col-start-2 grid grid-cols-[minmax(10rem,auto)_4fr]">
              <span />
              <span>
                <RenderMarkdown :content="error.message" />
              </span>
            </div>
          </div>
        </template>
      </div>

      <div
        v-if="runtimeErrorWorkflows.length > 0"
        class="flex flex-col gap-y-4"
        :class="hasHardParseErrors(pipeline) ? 'order-2' : 'order-1'"
      >
        <span class="text-lg font-bold">{{ $t('repo.pipeline.runtime_errors') }}</span>
        <div
          v-for="workflow in runtimeErrorWorkflows"
          :key="workflow.id"
          class="grid grid-cols-[minmax(10rem,auto)_3fr]"
        >
          <span class="flex items-start gap-x-2">
            <Icon name="alert" class="text-wp-error-100 my-1 shrink-0" />
            <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
            <span class="my-1">
              <code>{{ workflow.name }}</code>
            </span>
          </span>
          <pre class="code-box break-words whitespace-pre-wrap">{{ workflow.error }}</pre>
        </div>
      </div>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import DocsLink from '~/components/atomic/DocsLink.vue';
import Icon from '~/components/atomic/Icon.vue';
import RenderMarkdown from '~/components/atomic/RenderMarkdown.vue';
import Panel from '~/components/layout/Panel.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { PipelineError } from '~/lib/api/types';
import { hasHardParseErrors, pipelineHasErrorsToShow, workflowsWithErrors } from '~/lib/pipeline';

const repo = requiredInject('repo');
const pipeline = requiredInject('pipeline');

const runtimeErrorWorkflows = computed(() => workflowsWithErrors(pipeline.value));

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

const { t } = useI18n();
useWPTitle(
  computed(() => [
    pipelineHasErrorsToShow(pipeline.value) ? t('repo.pipeline.errors') : t('repo.pipeline.warnings'),
    t('repo.pipeline.pipeline', { pipelineId: pipeline.value.number }),
    repo.value.full_name,
  ]),
);
</script>
