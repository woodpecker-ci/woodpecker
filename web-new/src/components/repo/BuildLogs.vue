<template>
  <div v-if="build" class="bg-gray-700 p-4">
    <div v-for="(logLine, key) in logLines" :key="logLine" class="flex items-center">
      <div class="text-gray-500 text-sm w-4">{{ key + 1 }}</div>
      <div class="ml-6" v-html="logLine" />
      <div class="ml-auto text-gray-500 text-sm">{{ (key + 1) * 10 }}s</div>
    </div>
    <div v-if="exitCode !== undefined" class="text-gray-500 text-sm mt-4 ml-10">Exit code {{ exitCode }}</div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, PropType, ref, toRef } from 'vue';
import { Build } from '~/lib/api/types';
import AnsiConvert from 'ansi-to-html';
import useApiClient from '~/compositions/useApiClient';

export default defineComponent({
  name: 'BuildLogs',

  components: {},

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup(props) {
    const build = toRef(props, 'build');

    const apiClient = useApiClient();

    var ansiConvert = new AnsiConvert();
    const logLinesAnsi = ref<string[] | undefined>();
    const logLines = computed(() => logLinesAnsi.value?.map((logLine) => ansiConvert.toHtml(logLine)));
    const exitCode = ref<number | undefined>(0);

    onMounted(async () => {
      logLinesAnsi.value = [
        '\x1b[30mblack\x1b[37mwhite1',
        '\x1b[30mblack\x1b[37mwhite2',
        '\x1b[30mblack\x1b[37mwhite3',
        '\x1b[30mblack\x1b[37mwhite4',
      ];
    });

    return { logLines, build, exitCode };
  },
});
</script>
