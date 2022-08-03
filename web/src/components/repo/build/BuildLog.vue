<template>
  <div v-if="build" class="flex flex-col pt-10 md:pt-0">
    <div
      class="fixed top-0 left-0 w-full md:hidden flex px-4 py-2 bg-gray-600 dark:bg-dark-gray-800 text-gray-50"
      @click="$emit('update:proc-id', null)"
    >
      <span>{{ proc?.name }}</span>
      <Icon name="close" class="ml-auto" />
    </div>

    <div
      class="flex flex-grow flex-col bg-gray-300 dark:bg-dark-gray-700 md:m-2 md:mt-0 md:rounded-md overflow-hidden"
      @mouseover="showActions = true"
      @mouseleave="showActions = false"
    >
      <div v-show="showActions" class="absolute top-0 right-0 z-50 mt-2 mr-4 hidden md:flex">
        <Button
          v-if="proc?.end_time !== undefined"
          :is-loading="downloadInProgress"
          :title="$t('repo.build.actions.log_download')"
          start-icon="download"
          @click="download"
        />
      </div>

      <div
        v-show="loadedLogs"
        ref="consoleElement"
        class="w-full max-w-full flex-auto flex-grow p-2 overflow-x-hidden overflow-y-auto"
      >
        <div class="table">
          <div v-for="l in log" :key="l.line" class="table-row whitespace-pre-wrap font-mono">
            <span class="text-gray-500 table-cell whitespace-nowrap select-none pl-2 pr-2 align-top text-right">
              {{ l.line }}
            </span>
            <!-- eslint-disable-next-line vue/no-v-html -->
            <span class="table-cell align-top text-color whitespace-pre-wrap break-words w-[100%]" v-html="l.text" />
            <span class="text-gray-500 table-cell whitespace-nowrap select-none pl-2 pr-2 align-top text-right">
              {{ l.time }}
            </span>
          </div>
        </div>
      </div>

      <div class="m-auto text-xl text-color">
        <span v-if="proc?.error" class="text-red-400">{{ proc.error }}</span>
        <span v-else-if="proc?.state === 'skipped'" class="text-red-400">{{ $t('repo.build.actions.canceled') }}</span>
        <span v-else-if="!proc?.start_time">{{ $t('repo.build.step_not_started') }}</span>
        <div v-else-if="!loadedLogs">{{ $t('repo.build.loading') }}</div>
      </div>

      <div
        v-if="proc?.end_time !== undefined"
        :class="proc.exit_code == 0 ? 'dark:text-lime-400 text-lime-700' : 'dark:text-red-400 text-red-600'"
        class="w-full bg-gray-200 dark:bg-dark-gray-800 text-md p-4"
      >
        {{ $t('repo.build.exit_code', { exitCode: proc.exit_code }) }}
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import '~/style/console.css';

import AnsiUp from 'ansi_up';
import { computed, defineComponent, inject, nextTick, onMounted, PropType, Ref, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { Build, Repo } from '~/lib/api/types';
import { findProc, isProcFinished, isProcRunning } from '~/utils/helpers';

type LogLine = {
  line: number;
  text: string;
  time: string;
};

export default defineComponent({
  name: 'BuildLog',

  components: { Icon, Button },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },

    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    procId: {
      type: Number,
      required: true,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:proc-id': (procId: number | null) => true,
  },

  setup(props) {
    const notifications = useNotifications();
    const i18n = useI18n();
    const build = toRef(props, 'build');
    const procId = toRef(props, 'procId');
    const repo = inject<Ref<Repo>>('repo');
    const apiClient = useApiClient();

    const loadedProcSlug = ref<string>();
    const procSlug = computed(() => `${repo?.value.owner} - ${repo?.value.name} - ${build.value.id} - ${procId.value}`);
    const proc = computed(() => build.value && findProc(build.value.procs || [], procId.value));
    const stream = ref<EventSource>();
    const log = ref<LogLine[]>([]);
    const consoleElement = ref<Element>();

    const loadedLogs = ref(true);
    const autoScroll = ref(true); // TODO
    const showActions = ref(false);
    const downloadInProgress = ref(false);
    const ansiUp = ref(new AnsiUp());
    const maxLineCount = 500;
    ansiUp.value.use_classes = true;

    let logBatch: LogLine[] = [];
    let timer: number | undefined;

    function write(lines: LogLine[]) {
      let lastLine = 0;
      if (log.value.length > 0) {
        lastLine = log.value[log.value.length - 1].line;
      }
      if (logBatch.length > 0 && logBatch[logBatch.length - 1].line > lastLine) {
        lastLine = logBatch[logBatch.length - 1].line;
      }
      for (let i = 0; i < lines.length; i += 1) {
        const line = lines[i];
        if (line.line > lastLine) {
          logBatch.push({ ...line, text: ansiUp.value.ansi_to_html(line.text) });
        }
      }
      if (logBatch.length > maxLineCount) {
        logBatch.splice(0, logBatch.length - maxLineCount);
      }
    }

    function flush(): boolean {
      const b = logBatch.splice(0);
      if (b.length === 0) {
        return false;
      }
      if (b.length >= maxLineCount) {
        log.value = b.splice(0);
        return true;
      }
      if (log.value.length + b.length > maxLineCount) {
        log.value.splice(0, log.value.length + b.length - maxLineCount);
      }
      log.value.push(...b);
      return true;
    }

    function scrollDown() {
      nextTick(() => {
        if (!consoleElement.value) {
          return;
        }
        consoleElement.value.scrollTop = consoleElement.value.scrollHeight;
      });
    }

    async function download() {
      if (!repo?.value || !build.value || !proc.value) {
        throw new Error('The repository, build or proc was undefined');
      }
      let logs;
      try {
        downloadInProgress.value = true;
        logs = await apiClient.getLogs(repo.value.owner, repo.value.name, build.value.number, proc.value.pid);
      } catch (e) {
        notifications.notifyError(e, i18n.t('repo.build.log_download_error'));
        return;
      } finally {
        downloadInProgress.value = false;
      }
      const fileURL = window.URL.createObjectURL(
        new Blob([logs.map((line) => line.out).join('')], {
          type: 'text/plain',
        }),
      );
      const fileLink = document.createElement('a');

      fileLink.href = fileURL;
      fileLink.setAttribute(
        'download',
        `${repo.value.owner}-${repo.value.name}-${build.value.number}-${proc.value.name}.log`,
      );
      document.body.appendChild(fileLink);

      fileLink.click();
      document.body.removeChild(fileLink);
      window.URL.revokeObjectURL(fileURL);
    }

    async function loadLogs() {
      if (loadedProcSlug.value === procSlug.value) {
        return;
      }
      loadedProcSlug.value = procSlug.value;
      loadedLogs.value = false;
      log.value = [];
      logBatch = [];
      ansiUp.value = new AnsiUp();
      ansiUp.value.use_classes = true;
      if (timer) {
        window.clearTimeout(timer);
      }

      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      if (stream.value) {
        stream.value.close();
      }

      // we do not have logs for skipped jobs
      if (
        !repo.value ||
        !build.value ||
        !proc.value ||
        proc.value.state === 'skipped' ||
        proc.value.state === 'killed'
      ) {
        return;
      }

      if (isProcFinished(proc.value)) {
        const logs = await apiClient.getLogs(repo.value.owner, repo.value.name, build.value.number, proc.value.pid);
        write(
          logs
            .slice(Math.max(logs.length, 0) - maxLineCount, logs.length) // TODO: think about way to support lazy-loading more than last 300 logs (#776))
            .map((l) => ({
              line: l.pos,
              text: l.out,
              time: l.time ? `${l.time}s` : '',
            })),
        );
        flush();
        loadedLogs.value = true;
      }

      if (isProcRunning(proc.value)) {
        timer = window.setInterval(() => {
          if (flush() && autoScroll.value) {
            scrollDown();
          }
        }, 500);

        // load stream of parent process (which receives all child processes logs)
        // TODO: change stream to only send data of single child process
        stream.value = apiClient.streamLogs(
          repo.value.owner,
          repo.value.name,
          build.value.number,
          proc.value.ppid,
          (l) => {
            if (l?.proc !== proc.value?.name) {
              return;
            }
            loadedLogs.value = true;
            write([{ line: l.pos, text: l.out, time: l.time ? `${l.time}s` : '' }]);
          },
        );
      }
    }

    onMounted(async () => {
      loadLogs();
    });

    watch(procSlug, () => {
      loadLogs();
    });

    watch(proc, (oldProc, newProc) => {
      if (oldProc && oldProc.name === newProc?.name && oldProc?.end_time !== newProc?.end_time) {
        if (autoScroll.value) {
          scrollDown();
        }
      }
    });

    return { consoleElement, proc, log, loadedLogs, showActions, download, downloadInProgress };
  },
});
</script>
