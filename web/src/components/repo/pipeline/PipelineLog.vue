<template>
  <div v-if="pipeline" class="flex flex-col pt-10 md:pt-0">
    <div
      class="flex flex-grow flex-col code-box shadow !p-0 !rounded-none md:mt-0 !md:rounded-md overflow-hidden"
      @mouseover="showActions = true"
      @mouseleave="showActions = false"
    >
      <div class="<md:fixed <md:top-0 <md:left-0 flex flex-row items-center w-full bg-wp-code-100 px-4 py-2">
        <span class="text-base font-bold text-wp-code-text-alt-100">
          <span class="<md:hidden">{{ $t('repo.pipeline.log_title') }}</span>
          <span class="md:hidden">{{ step?.name }}</span>
        </span>

        <div class="flex flex-row items-center ml-auto gap-x-2">
          <IconButton
            v-if="step?.finished !== undefined && hasLogs"
            :is-loading="downloadInProgress"
            :title="$t('repo.pipeline.actions.log_download')"
            class="!hover:bg-white !hover:bg-opacity-10"
            icon="download"
            @click="download"
          />
          <IconButton
            v-if="step?.finished !== undefined && hasLogs && hasPushPermission"
            :title="$t('repo.pipeline.actions.log_delete')"
            class="!hover:bg-white !hover:bg-opacity-10"
            icon="trash"
            @click="deleteLogs"
          />
          <IconButton
            v-if="step?.finished === undefined"
            :title="
              autoScroll ? $t('repo.pipeline.actions.log_auto_scroll_off') : $t('repo.pipeline.actions.log_auto_scroll')
            "
            class="!hover:bg-white !hover:bg-opacity-10"
            :icon="autoScroll ? 'auto-scroll' : 'auto-scroll-off'"
            @click="autoScroll = !autoScroll"
          />
          <IconButton
            class="!hover:bg-white !hover:bg-opacity-10 !md:hidden"
            icon="close"
            @click="$emit('update:step-id', null)"
          />
        </div>
      </div>

      <div
        v-show="hasLogs && loadedLogs && (log?.length || 0) > 0"
        ref="consoleElement"
        class="w-full max-w-full grid grid-cols-[min-content,minmax(0,1fr),min-content] p-4 auto-rows-min flex-grow overflow-x-hidden overflow-y-auto text-xs md:text-sm"
      >
        <div v-for="line in log" :key="line.index" class="contents font-mono">
          <a
            :id="`L${line.number}`"
            :href="`#L${line.number}`"
            class="text-wp-code-text-alt-100 whitespace-nowrap select-none text-right pl-2 pr-6"
            :class="{
              'bg-opacity-40 dark:bg-opacity-50 bg-red-600 dark:bg-red-800': line.type === 'error',
              'bg-opacity-40 dark:bg-opacity-50 bg-yellow-600 dark:bg-yellow-800': line.type === 'warning',
              'bg-opacity-30 bg-blue-600': isSelected(line),
              underline: isSelected(line),
            }"
          >
            {{ line.number }}
          </a>
          <!-- eslint-disable vue/no-v-html -->
          <span
            class="align-top whitespace-pre-wrap break-words"
            :class="{
              'bg-opacity-40 dark:bg-opacity-50 bg-10.168.64.121-600 dark:bg-red-800': line.type === 'error',
              'bg-opacity-40 dark:bg-opacity-50 bg-yellow-600 dark:bg-yellow-800': line.type === 'warning',
              'bg-opacity-30 bg-blue-600': isSelected(line),
            }"
            v-html="line.text"
          />
          <!-- eslint-enable vue/no-v-html -->
          <span
            class="text-wp-code-text-alt-100 whitespace-nowrap select-none text-right pr-1"
            :class="{
              'bg-opacity-40 dark:bg-opacity-50 bg-red-600 dark:bg-red-800': line.type === 'error',
              'bg-opacity-40 dark:bg-opacity-50 bg-yellow-600 dark:bg-yellow-800': line.type === 'warning',
              'bg-opacity-30 bg-blue-600': isSelected(line),
            }"
          >
            {{ formatTime(line.time) }}
          </span>
        </div>
      </div>

      <div class="m-auto text-xl text-wp-text-alt-100">
        <span v-if="step?.state === 'skipped'">{{ $t('repo.pipeline.actions.canceled') }}</span>
        <span v-else-if="!step?.started">{{ $t('repo.pipeline.step_not_started') }}</span>
        <div v-else-if="!loadedLogs">{{ $t('repo.pipeline.loading') }}</div>
        <div v-else-if="log?.length === 0">{{ $t('repo.pipeline.no_logs') }}</div>
      </div>

      <div
        v-if="step?.finished !== undefined"
        class="flex items-center w-full bg-wp-code-100 text-md text-wp-code-text-alt-100 p-4 font-bold"
      >
        <PipelineStatusIcon :status="step.state" class="!h-4 !w-4" />
        <span v-if="step?.error" class="px-2">{{ step.error }}</span>
        <span v-else class="px-2">{{ $t('repo.pipeline.exit_code', { exitCode: step.exit_code }) }}</span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import '~/style/console.css';

import { useStorage } from '@vueuse/core';
import { AnsiUp } from 'ansi_up';
import { decode } from 'js-base64';
import { debounce } from 'lodash';
import { computed, inject, nextTick, onBeforeUnmount, onMounted, ref, toRef, watch, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import IconButton from '~/components/atomic/IconButton.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import type { Pipeline, Repo, RepoPermissions } from '~/lib/api/types';
import { findStep, isStepFinished, isStepRunning } from '~/utils/helpers';

interface LogLine {
  index: number;
  number: number;
  text?: string;
  time?: number;
  type: 'error' | 'warning' | null;
}

const props = defineProps<{
  pipeline: Pipeline;
  stepId: number;
}>();

defineEmits<{
  (event: 'update:step-id', stepId: number | null): true;
}>();

const notifications = useNotifications();
const i18n = useI18n();
const pipeline = toRef(props, 'pipeline');
const stepId = toRef(props, 'stepId');
const repo = inject<Ref<Repo>>('repo');
const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
const apiClient = useApiClient();
const route = useRoute();

const loadedStepSlug = ref<string>();
const stepSlug = computed(() => `${repo?.value.owner} - ${repo?.value.name} - ${pipeline.value.id} - ${stepId.value}`);
const step = computed(() => pipeline.value && findStep(pipeline.value.workflows || [], stepId.value));
const stream = ref<EventSource>();
const log = ref<LogLine[]>();
const consoleElement = ref<Element>();

const loadedLogs = computed(() => !!log.value);
const hasLogs = computed(
  () =>
    // we do not have logs for skipped steps
    repo?.value && pipeline.value && step.value && step.value.state !== 'skipped',
);
const autoScroll = useStorage('log-auto-scroll', false);
const showActions = ref(false);
const downloadInProgress = ref(false);
const ansiUp = ref(new AnsiUp());
ansiUp.value.use_classes = true;
const logBuffer = ref<LogLine[]>([]);

const maxLineCount = 5000; // TODO(2653): set back to 500 and implement lazy-loading support
const hasPushPermission = computed(() => repoPermissions?.value?.push);

function isSelected(line: LogLine): boolean {
  return route.hash === `#L${line.number}`;
}

function formatTime(time?: number): string {
  return time === undefined ? '' : `${time}s`;
}

function processText(text: string): string {
  const urlRegex = /https?:\/\/\S+/g;
  let txt = ansiUp.value.ansi_to_html(`${decode(text)}\n`);
  txt = txt.replace(
    urlRegex,
    (url) => `<a href="${url}" target="_blank" rel="noopener noreferrer" class="underline">${url}</a>`,
  );
  return txt;
}

function writeLog(line: Partial<LogLine>) {
  logBuffer.value.push({
    index: line.index ?? 0,
    number: (line.index ?? 0) + 1,
    text: processText(line.text ?? ''),
    time: line.time ?? 0,
    type: null, // TODO: implement way to detect errors and warnings
  });
}

function scrollDown() {
  nextTick(() => {
    if (!consoleElement.value) {
      return;
    }
    consoleElement.value.scrollTop = consoleElement.value.scrollHeight;
  });
}

const flushLogs = debounce((scroll: boolean) => {
  let buffer = logBuffer.value.slice(-maxLineCount);
  logBuffer.value = [];

  if (buffer.length === 0) {
    if (!log.value) {
      log.value = [];
    }
    return;
  }

  // append old logs lines
  if (buffer.length < maxLineCount && log.value) {
    buffer = [...log.value.slice(-(maxLineCount - buffer.length)), ...buffer];
  }

  // deduplicate repeating times
  buffer = buffer.reduce(
    (acc, line) => ({
      lastTime: line.time ?? 0,
      lines: [
        ...acc.lines,
        {
          ...line,
          time: acc.lastTime === line.time ? undefined : line.time,
        },
      ],
    }),
    { lastTime: -1, lines: [] as LogLine[] },
  ).lines;

  log.value = buffer;

  if (route.hash.length > 0) {
    nextTick(() => document.getElementById(route.hash.substring(1))?.scrollIntoView());
  } else if (scroll && autoScroll.value) {
    scrollDown();
  }
}, 500);

async function download() {
  if (!repo?.value || !pipeline.value || !step.value) {
    throw new Error('The repository, pipeline or step was undefined');
  }
  let logs;
  try {
    downloadInProgress.value = true;
    logs = await apiClient.getLogs(repo.value.id, pipeline.value.number, step.value.id);
  } catch (e) {
    notifications.notifyError(e as Error, i18n.t('repo.pipeline.log_download_error'));
    return;
  } finally {
    downloadInProgress.value = false;
  }
  const fileURL = window.URL.createObjectURL(
    new Blob([logs.map((line) => decode(line.data ?? '')).join('\n')], {
      type: 'text/plain',
    }),
  );
  const fileLink = document.createElement('a');

  fileLink.href = fileURL;
  fileLink.setAttribute(
    'download',
    `${repo.value.owner}-${repo.value.name}-${pipeline.value.number}-${step.value.name}.log`,
  );
  document.body.appendChild(fileLink);

  fileLink.click();
  document.body.removeChild(fileLink);
  window.URL.revokeObjectURL(fileURL);
}

async function loadLogs() {
  if (loadedStepSlug.value === stepSlug.value) {
    return;
  }

  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  log.value = undefined;
  logBuffer.value = [];
  ansiUp.value = new AnsiUp();
  ansiUp.value.use_classes = true;

  stream.value?.close();

  if (!hasLogs.value || !step.value) {
    return;
  }

  if (isStepFinished(step.value)) {
    loadedStepSlug.value = stepSlug.value;
    const logs = await apiClient.getLogs(repo.value.id, pipeline.value.number, step.value.id);
    logs?.forEach((line) => writeLog({ index: line.line, text: line.data, time: line.time }));
    flushLogs(false);
  } else if (step.value.state === 'pending' || isStepRunning(step.value)) {
    loadedStepSlug.value = stepSlug.value;
    stream.value = apiClient.streamLogs(repo.value.id, pipeline.value.number, step.value.id, (line) => {
      writeLog({ index: line.line, text: line.data, time: line.time });
      flushLogs(true);
    });
  }
}

async function deleteLogs() {
  if (!repo?.value || !pipeline.value || !step.value) {
    throw new Error('The repository, pipeline or step was undefined');
  }

  // TODO: use proper dialog (copy-pasted from web/src/components/secrets/SecretList.vue:deleteSecret)
  // eslint-disable-next-line no-alert
  if (!confirm(i18n.t('repo.pipeline.log_delete_confirm'))) {
    return;
  }

  try {
    await apiClient.deleteLogs(repo.value.id, pipeline.value.number, step.value.id);
    log.value = [];
  } catch (e) {
    notifications.notifyError(e as Error, i18n.t('repo.pipeline.log_delete_error'));
  }
}

onMounted(async () => {
  await loadLogs();
});

onBeforeUnmount(() => {
  stream.value?.close();
});

watch(stepSlug, async () => {
  await loadLogs();
});

watch(step, async (newStep, oldStep) => {
  if (oldStep?.name === newStep?.name) {
    if (oldStep?.finished !== newStep?.finished && autoScroll.value) {
      scrollDown();
    }

    if (oldStep?.state !== newStep?.state) {
      await loadLogs();
    }
  }
});
</script>
