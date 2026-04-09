<template>
  <div
    v-if="pipeline"
    class="fixed top-0 left-0 flex h-full w-full flex-col pt-10 md:pt-0"
    :class="{
      'md:absolute': !fullscreen,
    }"
  >
    <div
      class="code-box-log flex grow flex-col overflow-hidden p-0! md:mt-0"
      :class="{
        'md:rounded-md!': !fullscreen,
      }"
      @mouseover="showActions = true"
      @mouseleave="showActions = false"
    >
      <div
        class="bg-wp-code-100 fixed top-0 left-0 flex w-full flex-row items-center px-4 py-2 md:relative md:top-auto md:left-auto"
      >
        <span class="text-wp-code-text-alt-100 text-base font-bold">
          <span class="md:display-unset hidden">{{ $t('repo.pipeline.log_title') }}</span>
          <span class="md:hidden">{{ step?.name }}</span>
        </span>

        <div class="ml-auto flex flex-row items-center gap-x-2">
          <IconButton
            :title="fullscreen ? $t('exit_fullscreen') : $t('fullscreen')"
            class="hidden! hover:bg-white/10! md:flex!"
            :icon="fullscreen ? 'exit-fullscreen' : 'fullscreen'"
            @click="fullscreen = !fullscreen"
          />
          <IconButton
            v-if="step?.finished !== undefined && hasLogs"
            :is-loading="downloadInProgress"
            :title="$t('repo.pipeline.actions.log_download')"
            class="hover:bg-white/10!"
            icon="download"
            @click="download"
          />
          <IconButton
            v-if="step?.finished !== undefined && hasLogs && hasPushPermission"
            :title="$t('repo.pipeline.actions.log_delete')"
            class="hover:bg-white/10!"
            icon="trash"
            @click="deleteLogs"
          />
          <IconButton
            v-if="step?.finished === undefined"
            :title="
              autoScroll ? $t('repo.pipeline.actions.log_auto_scroll_off') : $t('repo.pipeline.actions.log_auto_scroll')
            "
            class="hover:bg-white/10!"
            :icon="autoScroll ? 'auto-scroll' : 'auto-scroll-off'"
            @click="autoScroll = !autoScroll"
          />
          <template v-if="hasGroupedLogs">
            <div class="border-wp-background-400 dark:border-wp-background-100 mx-1 h-5 border-l" />
            <IconButton
              :title="$t('repo.pipeline.actions.expand_all')"
              class="hover:bg-white/10!"
              icon="expand-all"
              @click="expandAll"
            />
            <IconButton
              :title="$t('repo.pipeline.actions.collapse_all')"
              class="hover:bg-white/10!"
              icon="collapse-all"
              @click="collapseAll"
            />
            <IconButton class="hover:bg-white/10! md:hidden!" icon="close" @click="$emit('update:step-id', null)" />
          </template>
        </div>
      </div>

      <div
        v-show="hasLogs && loadedLogs && (log?.length || 0) > 0"
        ref="consoleElement"
        class="grid w-full max-w-full grow scroll-pt-8 auto-rows-min grid-cols-[min-content_minmax(0,1fr)_min-content] overflow-x-hidden overflow-y-auto p-4 text-xs md:text-sm"
      >
        <div v-for="group in groupedLogs" :key="group.id" class="contents">
          <div
            v-if="group.isActualCommand"
            class="sticky -top-4 z-10 col-span-3 my-1 flex cursor-pointer items-center rounded-sm px-2 py-1 font-mono text-sm shadow-xs"
            :class="[group.command && isSelected(group.command) ? 'bg-blue-900' : 'bg-wp-code-100']"
            @click="toggleGroup(group.id)"
          >
            <Icon
              name="chevron-right"
              class="mr-2 transition-transform"
              :class="{
                'rotate-90': !collapsedCommands.has(group.id),
                invisible: group.lines.length === 0,
              }"
            />
            <a
              :id="`L${group.command?.number}`"
              :href="`#L${group.command?.number}`"
              class="text-wp-code-text-alt-100 mr-4 select-none"
              :class="{
                underline: group.command && isSelected(group.command),
              }"
            >
              {{ group.command?.number }}
            </a>
            <!-- eslint-disable vue/no-v-html -->
            <span v-if="group.command" class="flex-1 truncate" v-html="group.command.text" />
          </div>

          <template v-if="!collapsedCommands.has(group.id)">
            <div v-for="line in group.lines" :key="line.index" class="contents font-mono">
              <a
                :id="`L${line.number}`"
                :href="`#L${line.number}`"
                class="text-wp-code-text-alt-100 pr-6 pl-2 text-right whitespace-nowrap select-none"
                :class="{
                  'bg-red-600/40 dark:bg-red-800/50': line.type === 'error',
                  'bg-yellow-600/40 dark:bg-yellow-800/50': line.type === 'warning',
                  'bg-blue-600/30': isSelected(line),
                  underline: isSelected(line),
                }"
              >
                {{ line.number }}
              </a>
              <!-- eslint-disable vue/no-v-html -->
              <span
                class="wrap-break-words align-top whitespace-pre-wrap"
                :class="{
                  'bg-red-600/40 dark:bg-red-800/50': line.type === 'error',
                  'bg-yellow-600/40 dark:bg-yellow-800/50': line.type === 'warning',
                  'bg-blue-600/30': isSelected(line),
                }"
                v-html="line.text"
              />
              <!-- eslint-enable vue/no-v-html -->
              <span
                class="text-wp-code-text-alt-100 pr-1 text-right whitespace-nowrap select-none"
                :class="{
                  'bg-red-600/40 dark:bg-red-800/50': line.type === 'error',
                  'bg-yellow-600/40 dark:bg-yellow-800/50': line.type === 'warning',
                  'bg-blue-600/30': isSelected(line),
                }"
              >
                {{ formatTime(line.time) }}
              </span>
            </div>
          </template>
        </div>
      </div>

      <div class="text-wp-text-alt-100 m-auto text-xl">
        <span v-if="step?.state === 'canceled'">{{ $t('repo.pipeline.actions.canceled') }}</span>
        <span v-else-if="step?.state === 'skipped'">{{ $t('repo.pipeline.actions.skipped') }}</span>
        <span v-else-if="!step?.started">{{ $t('repo.pipeline.step_not_started') }}</span>
        <div v-else-if="!loadedLogs">{{ $t('repo.pipeline.loading') }}</div>
        <div v-else-if="log?.length === 0">{{ $t('repo.pipeline.no_logs') }}</div>
      </div>

      <div
        v-if="step?.finished !== undefined"
        class="text-md bg-wp-code-100 text-wp-code-text-alt-100 flex w-full items-center p-4 font-bold"
      >
        <PipelineStatusIcon :status="step.state" class="h-4! w-4!" />
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
import { computed, nextTick, onBeforeUnmount, onMounted, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import useApiClient from '~/compositions/useApiClient';
import useConfig from '~/compositions/useConfig';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import type { Pipeline, PipelineConfig, PipelineStep, PipelineWorkflow } from '~/lib/api/types';
import { debounce } from '~/lib/utils';

interface LogLine {
  index: number;
  number: number;
  text?: string;
  rawText?: string;
  time?: number;
  type: 'error' | 'warning' | null;
}

interface LogBlock {
  command: LogLine | null;
  lines: LogLine[];
  id: number;
  isActualCommand: boolean;
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
const repo = requiredInject('repo');
const repoPermissions = requiredInject('repo-permissions');
const pipelineConfigs = requiredInject('pipeline-configs');
const apiClient = useApiClient();
const route = useRoute();

const loadedStepSlug = ref<string>();
const stepSlug = computed(() => `${repo?.value.owner} - ${repo?.value.name} - ${pipeline.value.id} - ${stepId.value}`);
const step = computed(() => pipeline.value && findStep(pipeline.value.workflows || [], stepId.value));
const stream = ref<EventSource>();
const log = ref<LogLine[]>();
const consoleElement = ref<Element>();
const fullscreen = ref(false);

const loadedLogs = computed(() => !!log.value);
const hasLogs = computed(
  () =>
    // we do not have logs for skipped/canceled steps
    repo?.value && pipeline.value && step.value && step.value.state !== 'skipped' && step.value.state !== 'canceled',
);
const autoScroll = useStorage('woodpecker:log-auto-scroll', true);
const showActions = ref(false);
const downloadInProgress = ref(false);
const ansiUp = ref(new AnsiUp());
ansiUp.value.use_classes = true;
const logBuffer = ref<LogLine[]>([]);

const config = useConfig();

const maxLineCount = config.maxPipelineLogLineCount; // TODO(2653): implement lazy-loading support
const hasPushPermission = computed(() => repoPermissions?.value?.push);

const collapsedCommands = ref(new Set<number>());

const commandRegex = /^\s*-\s(.+)$/gm;
const specialCharsRegex = /[.*+?^${}()|[\]\\]/g;
const matrixVariableRegex = /\\\$(\\\{\w+\\\})/g;

const knownCommandMatchers = computed(() => {
  if (!pipelineConfigs.value) return [];
  const patterns: RegExp[] = [];
  pipelineConfigs.value.forEach((config: PipelineConfig) => {
    const decoded = decode(config.data);
    const matches = decoded.matchAll(commandRegex);
    for (const match of matches) {
      const rawCommand = match[1].trim();
      // Replace matrix variable ${VAR} with a wildcard match (non-greedy)
      const patternString = rawCommand
        .replace(specialCharsRegex, '\\$&') // escape all
        .replace(matrixVariableRegex, '.*'); // match ${VAR}

      patterns.push(new RegExp(`^${patternString}$`));
    }
  });
  return patterns;
});

const groupedLogs = computed(() => {
  if (!log.value) return [];

  if (!pipelineConfigs.value || pipelineConfigs.value.length === 0) {
    return [
      {
        id: 0,
        command: null,
        lines: log.value,
        isActualCommand: false,
      },
    ];
  }

  const blocks: LogBlock[] = [];
  let currentBlock: LogBlock | null = null;

  log.value.forEach((line) => {
    const trimmedText = (line.rawText || '').trim();

    let isCommand = false;
    if (trimmedText.startsWith('+ ')) {
      const cmdPart = trimmedText.slice(2).trim();
      isCommand = knownCommandMatchers.value.some((matcher) => matcher.test(cmdPart));
    }

    if (isCommand) {
      currentBlock = {
        command: line,
        lines: [],
        id: line.number,
        isActualCommand: true,
      };
      blocks.push(currentBlock);
    } else {
      if (!currentBlock) {
        currentBlock = {
          command: { number: 0, text: 'Initialization', type: null, index: -1 } as LogLine,
          lines: [],
          id: 0,
          isActualCommand: false,
        };
        blocks.push(currentBlock);
      }
      currentBlock.lines.push(line);
    }
  });

  return blocks;
});

const hasGroupedLogs = computed(() => {
  return groupedLogs.value.find((g) => g.isActualCommand);
});

const urlRegex = /https?:\/\/\S+/g;

function isScrolledToBottom(): boolean {
  if (!consoleElement.value) {
    return false;
  }
  // we use 5 as threshold
  return consoleElement.value.scrollHeight - consoleElement.value.scrollTop - consoleElement.value.clientHeight < 5;
}

function isSelected(line: LogLine): boolean {
  return route.hash === `#L${line.number}`;
}

function formatTime(time?: number): string {
  return time === undefined ? '' : `${time}s`;
}

function toggleGroup(id: number) {
  if (collapsedCommands.value.has(id)) {
    collapsedCommands.value.delete(id);
  } else {
    collapsedCommands.value.add(id);
  }
}

function expandAll() {
  collapsedCommands.value.clear();
}

function collapseAll() {
  const newSet = new Set<number>();
  groupedLogs.value.forEach((group) => {
    if (group.isActualCommand) {
      newSet.add(group.id);
    }
  });
  collapsedCommands.value = newSet;
}

function processText(text: string): string {
  let txt = ansiUp.value.ansi_to_html(`${decode(text)}\n`);
  txt = txt.replace(
    urlRegex,
    (url) => `<a href="${url}" target="_blank" rel="noopener noreferrer" class="underline">${url}</a>`,
  );
  return txt;
}

function writeLog(line: Partial<LogLine>) {
  const rawText = decode(line.text ?? '');
  logBuffer.value.push({
    index: line.index ?? 0,
    number: (line.index ?? 0) + 1,
    text: processText(line.text ?? ''),
    rawText,
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
  } else if (scroll && autoScroll.value && isScrolledToBottom()) {
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

  log.value = undefined;
  logBuffer.value = [];
  ansiUp.value = new AnsiUp();
  ansiUp.value.use_classes = true;

  stream.value?.close();

  if (!hasLogs.value || !step.value) {
    return;
  }

  if (step.value.state !== 'running' && step.value.state !== 'pending') {
    loadedStepSlug.value = stepSlug.value;
    const logs = await apiClient.getLogs(repo.value.id, pipeline.value.number, step.value.id);
    logs?.forEach((line) => writeLog({ index: line.line, text: line.data, time: line.time }));
    flushLogs(false);
  } else {
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

function findStep(workflows: PipelineWorkflow[], pid: number): PipelineStep | undefined {
  return workflows.reduce(
    (prev, workflow) => {
      const result = workflow.children.reduce(
        (prevChild, step) => {
          if (step.pid === pid) {
            return step;
          }

          return prevChild;
        },
        undefined as PipelineStep | undefined,
      );
      if (result) {
        return result;
      }

      return prev;
    },
    undefined as PipelineStep | undefined,
  );
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
    if (oldStep?.finished !== newStep?.finished && autoScroll.value && isScrolledToBottom()) {
      scrollDown();
    }

    if (oldStep?.state !== newStep?.state) {
      await loadLogs();
    }
  }
});

// When user click on a step, if the step has already finished running, show user the
// only the outline by collapse all log groups
watch(loadedLogs, async (isLoaded, wasLoaded) => {
  // Only trigger when transitioning from unloaded to loaded state
  if (isLoaded && !wasLoaded) {
    const isFinished = step.value && !['running', 'pending', 'started'].includes(step.value.state);
    if (isFinished) {
      // Wait for groupedLogs computed property to update
      await nextTick();
      collapseAll();
    }
  }
});

// If route hash contain line that is in a collapsed log group, expand it
watch(
  () => route.hash,
  (newHash) => {
    if (newHash.startsWith('#L')) {
      const lineNum = Number.parseInt(newHash.substring(2));
      const parentGroup = groupedLogs.value.find((g) => lineNum === g.id || g.lines.some((l) => l.number === lineNum));
      if (parentGroup && collapsedCommands.value.has(parentGroup.id)) {
        collapsedCommands.value.delete(parentGroup.id);
      }
    }
  },
  { immediate: true },
);
</script>
