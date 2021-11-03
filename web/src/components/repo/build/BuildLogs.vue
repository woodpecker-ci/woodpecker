<template>
  <div v-if="build" class="bg-gray-700 dark:bg-dark-gray-700 p-4">
    <div v-for="logLine in logLines" :key="logLine.pos" class="flex items-center">
      <div class="text-gray-500 text-sm w-4">{{ (logLine.pos || 0) + 1 }}</div>
      <!-- eslint-disable-next-line vue/no-v-html -->
      <div class="mx-4 text-gray-200 dark:text-gray-400" v-html="logLine.out" />
      <div class="ml-auto text-gray-500 text-sm">{{ logLine.time || 0 }}s</div>
    </div>
    <div v-if="proc?.end_time !== undefined" class="text-gray-500 text-sm mt-4 ml-8">
      exit code {{ proc.exit_code }}
    </div>
    <template v-if="!proc?.start_time" />
    <div class="text-gray-300 mx-auto">
      <span v-if="proc?.state === 'skipped'" class="text-orange-300 dark:text-orange-800"
        >This step has been canceled.</span
      >
      <span v-else-if="!proc?.start_time" class="dark:text-gray-500">This step hasn't started yet.</span>
    </div>
  </div>
</template>

<script lang="ts">
import AnsiConvert from 'ansi-to-html';
import { computed, defineComponent, inject, onBeforeUnmount, onMounted, PropType, Ref, toRef, watch } from 'vue';

import useBuildProc from '~/compositions/useBuildProc';
import { Build, Repo } from '~/lib/api/types';
import { findProc } from '~/utils/helpers';

export default defineComponent({
  name: 'BuildLogs',

  components: {},

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

  setup(props) {
    const build = toRef(props, 'build');
    const procId = toRef(props, 'procId');
    const repo = inject<Ref<Repo>>('repo');
    const buildProc = useBuildProc();

    const ansiConvert = new AnsiConvert();
    const logLines = computed(() => buildProc.logs.value?.map((l) => ({ ...l, out: ansiConvert.toHtml(l.out) })));
    const proc = computed(() => build.value && findProc(build.value.procs || [], procId.value));

    function loadBuildProc() {
      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      if (!repo.value || !build.value || !proc.value) {
        return;
      }

      buildProc.load(repo.value.owner, repo.value.name, build.value.number, proc.value);
    }

    onMounted(() => {
      loadBuildProc();
    });

    watch([repo, build, procId], () => {
      loadBuildProc();
    });

    onBeforeUnmount(() => {
      buildProc.unload();
    });

    return { logLines, proc };
  },
});
</script>
