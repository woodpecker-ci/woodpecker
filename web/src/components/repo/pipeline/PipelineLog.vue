<template>
  <div v-if="pipeline" class="flex flex-col pt-10 md:pt-0">
    <div
      class="fixed top-0 left-0 w-full md:hidden flex px-4 py-2 bg-gray-600 dark:bg-dark-gray-800 text-gray-50"
      @click="$emit('update:step-id', null)"
    >
      <span>{{ step?.name }}</span>
      <Icon name="close" class="ml-auto" />
    </div>

    <div
      class="flex flex-grow flex-col bg-white shadow dark:bg-dark-gray-700 md:m-2 md:mt-0 md:rounded-md overflow-hidden"
      @mouseover="showActions = true"
      @mouseleave="showActions = false"
    >
      <div v-show="showActions" class="absolute top-0 right-0 z-40 mt-2 mr-4 hidden md:flex">
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
        class="w-full max-w-full grid grid-cols-[min-content,1fr,min-content] auto-rows-min flex-grow p-2 gap-x-2 overflow-x-hidden overflow-y-auto"
      >
        <div v-for="line in log" :id="`L${line.index}`" :key="line.index" class="contents font-mono">
          <span class="text-gray-500 whitespace-nowrap select-none text-right">{{ line.index + 1 }}</span>
          <!-- eslint-disable-next-line vue/no-v-html -->
          <span class="align-top text-color whitespace-pre-wrap break-words" v-html="line.text" />
          <span class="text-gray-500 whitespace-nowrap select-none text-right">{{ formatTime(line.time) }}</span>
        </div>
      </div>

      <div class="m-auto text-xl text-color">
        <span v-if="step?.error" class="text-red-400">{{ step.error }}</span>
        <span v-else-if="step?.state === 'skipped'" class="text-red-400">{{
          $t('repo.pipeline.actions.canceled')
        }}</span>
        <span v-else-if="!step?.start_time">{{ $t('repo.pipeline.step_not_started') }}</span>
        <div v-else-if="!loadedLogs">{{ $t('repo.pipeline.loading') }}</div>
      </div>

      <div
        v-if="step?.end_time !== undefined"
        :class="step.exit_code == 0 ? 'dark:text-lime-400 text-lime-700' : 'dark:text-red-400 text-red-600'"
        class="w-full bg-gray-200 dark:bg-dark-gray-800 text-md p-4"
      >
        {{ $t('repo.pipeline.exit_code', { exitCode: step.exit_code }) }}
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

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { Pipeline, Repo } from '~/lib/api/types';
import { findStep, isStepFinished, isStepRunning } from '~/utils/helpers';

type LogLine = {
  index: number;
  text: string;
  time?: number;
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

const loadedStepSlug = ref<string>();
const stepSlug = computed(() => `${repo?.value.owner} - ${repo?.value.name} - ${pipeline.value.id} - ${stepId.value}`);
const step = computed(() => pipeline.value && findStep(pipeline.value.steps || [], stepId.value));
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

function formatTime(time?: number): string {
  return time === undefined ? '' : `${time}s`;
}

function writeLog(line: LogLine) {
  logBuffer.value.push({
    index: line.index ?? 0,
    text: ansiUp.value.ansi_to_html(line.text),
    time: line.time ?? 0,
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

  if (scroll && autoScroll.value) {
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
    logs = await apiClient.getLogs(repo.value.owner, repo.value.name, pipeline.value.number, step.value.id);
  } catch (e) {
    notifications.notifyError(e, i18n.t('repo.pipeline.log_download_error'));
    return;
  } finally {
    downloadInProgress.value = false;
  }
  const fileURL = window.URL.createObjectURL(
    new Blob([logs.map((line) => atob(line.data)).join('')], {
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
    const logs = await apiClient.getLogs(repo.value.owner, repo.value.name, pipeline.value.number, step.value.id);
    logs?.forEach((line) => writeLog({ index: line.line, text: atob(line.data), time: line.time }));
    flushLogs(false);
  }

  if (isStepRunning(step.value)) {
    stream.value = apiClient.streamLogs(
      repo.value.owner,
      repo.value.name,
      pipeline.value.number,
      step.value.id,
      (line) => {
        writeLog({ index: line.line, text: line.data, time: line.time });
        flushLogs(true);
      },
    );
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
