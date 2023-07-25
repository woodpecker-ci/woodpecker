<template>
  <div v-if="pipeline" class="flex flex-col pt-10 md:pt-0">
    <div
      class="fixed top-0 left-0 w-full md:hidden flex px-4 py-2 bg-wp-background-100 dark:bg-wp-background-200 text-wp-text-100"
      @click="$emit('update:step-id', null)"
    >
      <span>{{ step?.name }}</span>
      <Icon name="close" class="ml-auto" />
    </div>

    <div
      class="flex flex-grow flex-col code-box shadow !p-0 !rounded-none md:m-2 md:mt-0 !md:rounded-md overflow-hidden"
      @mouseover="showActions = true"
      @mouseleave="showActions = false"
    >
      <div v-show="showActions" class="absolute top-0 right-0 z-40 mt-4 mr-6 hidden md:flex">
        <Button
          v-if="step?.end_time !== undefined"
          :is-loading="downloadInProgress"
          :title="$t('repo.pipeline.actions.log_download')"
          start-icon="download"
          @click="download"
        />
        <Button
          v-if="step?.end_time === undefined"
          :title="
            autoScroll ? $t('repo.pipeline.actions.log_auto_scroll_off') : $t('repo.pipeline.actions.log_auto_scroll')
          "
          :start-icon="autoScroll ? 'auto-scroll' : 'auto-scroll-off'"
          @click="autoScroll = !autoScroll"
        />
      </div>

      <div
        v-show="hasLogs && loadedLogs"
        ref="consoleElement"
        class="w-full max-w-full grid grid-cols-[min-content,1fr,min-content] p-4 auto-rows-min flex-grow overflow-x-hidden overflow-y-auto"
      >
        <div v-for="line in log" :key="line.index" class="contents font-mono">
          <a
            :id="`L${line.number}`"
            :href="`#L${line.number}`"
            class="text-wp-text-alt-100 whitespace-nowrap select-none text-right pl-2 pr-6"
            :class="{
              'bg-opacity-40 dark:bg-opacity-50 bg-red-600 dark:bg-red-800': line.type === 'error',
              'bg-opacity-40 dark:bg-opacity-50 bg-yellow-600 dark:bg-yellow-800': line.type === 'warning',
              'bg-opacity-30 bg-blue-600': isSelected(line),
              underline: isSelected(line),
            }"
            >{{ line.number }}</a
          >
          <!-- eslint-disable vue/no-v-html -->
          <span
            class="align-top whitespace-pre-wrap break-words text-sm"
            :class="{
              'bg-opacity-40 dark:bg-opacity-50 bg-10.168.64.121-600 dark:bg-red-800': line.type === 'error',
              'bg-opacity-40 dark:bg-opacity-50 bg-yellow-600 dark:bg-yellow-800': line.type === 'warning',
              'bg-opacity-30 bg-blue-600': isSelected(line),
            }"
            v-html="line.text"
          />
          <!-- eslint-enable vue/no-v-html -->
          <span
            class="text-wp-text-alt-100 whitespace-nowrap select-none text-right pr-1"
            :class="{
              'bg-opacity-40 dark:bg-opacity-50 bg-red-600 dark:bg-red-800': line.type === 'error',
              'bg-opacity-40 dark:bg-opacity-50 bg-yellow-600 dark:bg-yellow-800': line.type === 'warning',
              'bg-opacity-30 bg-blue-600': isSelected(line),
            }"
            >{{ formatTime(line.time) }}</span
          >
        </div>
      </div>

      <div class="m-auto text-xl text-wp-text-alt-100">
        <span v-if="step?.error">{{ step.error }}</span>
        <span v-else-if="step?.state === 'skipped'">{{ $t('repo.pipeline.actions.canceled') }}</span>
        <span v-else-if="!step?.start_time">{{ $t('repo.pipeline.step_not_started') }}</span>
        <div v-else-if="!loadedLogs">{{ $t('repo.pipeline.loading') }}</div>
      </div>

      <div
        v-if="step?.end_time !== undefined"
        class="flex items-center w-full bg-wp-code-100 text-md text-wp-text-alt-100 p-4 font-bold"
      >
        <PipelineStatusIcon :status="step.state" class="!h-4 !w-4" />
        <span class="px-2">{{ $t('repo.pipeline.exit_code', { exitCode: step.exit_code }) }}</span>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import '~/style/console.css';

import { useStorage } from '@vueuse/core';
import AnsiUp from 'ansi_up';
import { debounce } from 'lodash';
import { computed, inject, nextTick, onMounted, Ref, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { Pipeline, Repo } from '~/lib/api/types';
import { findStep, isStepFinished, isStepRunning } from '~/utils/helpers';

type LogLine = {
  index: number;
  number: number;
  text: string;
  time?: number;
  type: 'error' | 'warning' | null;
};

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
    repo?.value && pipeline.value && step.value && step.value.state !== 'skipped' && step.value.state !== 'killed',
);
const autoScroll = useStorage('log-auto-scroll', false);
const showActions = ref(false);
const downloadInProgress = ref(false);
const ansiUp = ref(new AnsiUp());
ansiUp.value.use_classes = true;
const logBuffer = ref<LogLine[]>([]);

const maxLineCount = 500; // TODO: think about way to support lazy-loading more than last 300 logs (#776)

function isSelected(line: LogLine): boolean {
  return route.hash === `#L${line.number}`;
}

function formatTime(time?: number): string {
  return time === undefined ? '' : `${time}s`;
}

function writeLog(line: Partial<LogLine>) {
  logBuffer.value.push({
    index: line.index ?? 0,
    number: (line.index ?? 0) + 1,
    text: ansiUp.value.ansi_to_html(line.text ?? ''),
    time: line.time ?? 0,
    type: null, // TODO: implement way to detect errors and warnings
  });
}

// SOURCE: https://stackoverflow.com/questions/30106476/using-javascripts-atob-to-decode-base64-doesnt-properly-decode-utf-8-strings
function b64DecodeUnicode(str: string) {
  return decodeURIComponent(
    window
      .atob(str)
      .split('')
      .map((c) => `%${`00${c.charCodeAt(0).toString(16)}`.slice(-2)}`)
      .join(''),
  );
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
    notifications.notifyError(e, i18n.t('repo.pipeline.log_download_error'));
    return;
  } finally {
    downloadInProgress.value = false;
  }
  const fileURL = window.URL.createObjectURL(
    new Blob([logs.map((line) => b64DecodeUnicode(line.data)).join('')], {
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
  loadedStepSlug.value = stepSlug.value;
  log.value = undefined;
  logBuffer.value = [];
  ansiUp.value = new AnsiUp();
  ansiUp.value.use_classes = true;

  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  if (stream.value) {
    stream.value.close();
  }

  if (!hasLogs.value || !step.value) {
    return;
  }

  if (isStepFinished(step.value)) {
    const logs = await apiClient.getLogs(repo.value.id, pipeline.value.number, step.value.id);
    logs?.forEach((line) => writeLog({ index: line.line, text: b64DecodeUnicode(line.data), time: line.time }));
    flushLogs(false);
  }

  if (isStepRunning(step.value)) {
    stream.value = apiClient.streamLogs(repo.value.id, pipeline.value.number, step.value.id, (line) => {
      writeLog({ index: line.line, text: b64DecodeUnicode(line.data), time: line.time });
      flushLogs(true);
    });
  }
}

onMounted(async () => {
  loadLogs();
});

watch(stepSlug, () => {
  loadLogs();
});

watch(step, (oldStep, newStep) => {
  if (oldStep && oldStep.name === newStep?.name && oldStep?.end_time !== newStep?.end_time) {
    if (autoScroll.value) {
      scrollDown();
    }
  }
});
</script>
