<template>
  <div v-if="build" class="flex flex-col pt-10 md:pt-0">
    <div
      class="fixed top-0 left-0 w-full md:hidden flex px-4 py-2 bg-gray-600 dark:bg-dark-gray-800 text-gray-50"
      @click="$emit('update:proc-id', null)"
    >
      <span>{{ proc?.name }}</span>
      <Icon name="close" class="ml-auto" />
    </div>

    <div class="flex flex-grow flex-col bg-gray-300 dark:bg-dark-gray-700 md:m-2 md:mt-0 md:rounded-md overflow-hidden">
      <div v-show="loadedLogs" class="w-full flex-grow p-2">
        <div id="terminal" class="w-full h-full" />
      </div>

      <div class="m-auto text-xl text-gray-500 dark:text-gray-500">
        <span v-if="proc?.error" class="text-red-400">{{ proc.error }}</span>
        <span v-else-if="proc?.state === 'skipped'" class="text-red-400">{{ $t('repo.build.actions.canceled') }}</span>
        <span v-else-if="!proc?.start_time">{{ $t('repo.build.step_not_started') }}</span>
        <div v-else-if="!loadedLogs">{{ $t('repo.build.loading') }}</div>
      </div>

      <div
        v-if="proc?.end_time !== undefined"
        class="w-full bg-gray-400 dark:bg-dark-gray-800 text-gray-200 text-md p-4"
      >
        {{ $t('repo.build.exit_code', { exitCode: proc.exit_code }) }}
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import 'xterm/css/xterm.css';

import {
  computed,
  defineComponent,
  inject,
  nextTick,
  onBeforeUnmount,
  onMounted,
  PropType,
  Ref,
  ref,
  toRef,
  watch,
} from 'vue';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';

import Icon from '~/components/atomic/Icon.vue';
import useApiClient from '~/compositions/useApiClient';
import { useDarkMode } from '~/compositions/useDarkMode';
import { Build, Repo } from '~/lib/api/types';
import { findProc, isProcFinished, isProcRunning } from '~/utils/helpers';

export default defineComponent({
  name: 'BuildLog',

  components: { Icon },

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
    const build = toRef(props, 'build');
    const procId = toRef(props, 'procId');
    const repo = inject<Ref<Repo>>('repo');
    const apiClient = useApiClient();

    const loadedProcSlug = ref<string>();
    const procSlug = computed(() => `${repo?.value.owner} - ${repo?.value.name} - ${build.value.id} - ${procId.value}`);
    const proc = computed(() => build.value && findProc(build.value.procs || [], procId.value));
    const stream = ref<EventSource>();
    const term = ref(
      new Terminal({
        convertEol: true,
        disableStdin: true,
        theme: {
          cursor: 'transparent',
        },
      }),
    );
    const fitAddon = ref(new FitAddon());
    const loadedLogs = ref(true);
    const autoScroll = ref(true); // TODO

    async function loadLogs() {
      if (loadedProcSlug.value === procSlug.value) {
        return;
      }
      loadedProcSlug.value = procSlug.value;
      loadedLogs.value = false;
      term.value.reset();
      term.value.write('\x1b[?25l');

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
        term.value.write(
          logs
            .slice(Math.max(logs.length, 0) - 300, logs.length) // TODO: think about way to support lazy-loading more than last 300 logs (#776)
            .map((line) => `${(line.pos || 0).toString().padEnd(logs.length.toString().length)}  ${line.out}`)
            .join(''),
        );
        loadedLogs.value = true;
      }

      if (isProcRunning(proc.value)) {
        // load stream of parent process (which receives all child processes logs)
        // TODO: change stream to only send data of single child process
        stream.value = apiClient.streamLogs(
          repo.value.owner,
          repo.value.name,
          build.value.number,
          proc.value.ppid,
          (l) => {
            loadedLogs.value = true;
            term.value.write(l.out, () => {
              if (autoScroll.value) {
                term.value.scrollToBottom();
              }
            });
          },
        );
      }
    }

    function resize() {
      fitAddon.value.fit();
    }

    onMounted(async () => {
      term.value.loadAddon(fitAddon.value);
      term.value.loadAddon(new WebLinksAddon());

      await nextTick(() => {
        const element = document.getElementById('terminal');
        if (element === null) {
          throw new Error('Unexpected: "terminal" should be provided at this place');
        }
        term.value.open(element);
        fitAddon.value.fit();

        window.addEventListener('resize', resize);
      });

      loadLogs();
    });

    watch(procSlug, () => {
      loadLogs();
    });

    const { darkMode } = useDarkMode();
    watch(
      darkMode,
      () => {
        if (darkMode.value) {
          term.value.options = {
            theme: {
              background: '#303440', // dark-gray-700
              foreground: '#d3d3d3', // gray-...
            },
          };
        } else {
          term.value.options = {
            theme: {
              background: 'rgb(209,213,219)', // gray-300
              foreground: '#000',
              selection: '#000',
            },
          };
        }
      },
      { immediate: true },
    );

    onBeforeUnmount(() => {
      if (stream.value) {
        stream.value.close();
      }
      window.removeEventListener('resize', resize);
    });

    return { proc, loadedLogs };
  },
});
</script>
