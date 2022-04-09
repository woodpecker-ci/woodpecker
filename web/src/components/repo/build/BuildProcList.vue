<template>
  <div class="flex flex-col w-full md:w-3/12 text-gray-200 dark:text-gray-400 bg-gray-600 dark:bg-dark-gray-800">
    <div
      class="
        flex
        py-4
        px-2
        mx-2
        space-x-1
        justify-between
        flex-shrink-0
        text-gray-500
        border-b-1
        dark:border-dark-gray-600
      "
    >
      <div class="flex space-x-1 items-center flex-shrink-0">
        <div class="flex items-center"><img class="w-6" :src="build.author_avatar" /></div>
        <span>{{ build.author }}</span>
      </div>
      <div class="flex space-x-1 items-center min-w-0">
        <Icon v-if="build.event === 'push'" name="push" />
        <Icon v-if="build.event === 'deployment'" name="deployment" />
        <Icon v-else-if="build.event === 'tag'" name="tag" />
        <a
          v-else-if="build.event === 'pull_request'"
          class="flex items-center space-x-1 text-link min-w-0"
          :href="build.link_url"
          target="_blank"
        >
          <Icon name="pull_request" />
          <span class="truncate">{{ prettyRef }}</span>
        </a>
        <span v-if="build.event !== 'pull_request'" class="truncate">{{ build.branch }}</span>
      </div>
      <div class="flex items-center flex-shrink-0">
        <template v-if="build.event === 'pull_request'">
          <Icon name="commit" />
          <span>{{ build.commit.slice(0, 10) }}</span>
        </template>
        <a v-else class="text-link flex items-center" :href="build.link_url" target="_blank">
          <Icon name="commit" />
          <span>{{ build.commit.slice(0, 10) }}</span>
        </a>
      </div>
    </div>

    <div v-if="build.procs === undefined || build.procs.length === 0" class="m-auto mt-4">
      <span>No pipeline steps available!</span>
    </div>

    <div class="flex flex-grow relative min-h-0 overflow-y-auto">
      <div class="md:absolute top-0 left-0 w-full">
        <div v-for="proc in build.procs" :key="proc.id">
          <div class="p-4 pb-1 flex flex-wrap items-center justify-between">
            <span>{{ proc.name }}</span>
            <div v-if="proc.environ" class="text-xs">
              <div v-for="(value, key) in proc.environ" :key="key">
                <span
                  class="
                    pl-2
                    pr-1
                    py-0.5
                    bg-gray-800
                    dark:bg-gray-600
                    border-2 border-gray-800
                    dark:border-gray-600
                    rounded-l-full
                  "
                  >{{ key }}</span
                >
                <span class="pl-1 pr-2 py-0.5 border-2 border-gray-800 dark:border-gray-600 rounded-r-full">{{
                  value
                }}</span>
              </div>
            </div>
          </div>
          <div v-if="proc.children === undefined || proc.children.length === 0" class="text-center m-auto">
            <span>No pipeline steps available!</span>
          </div>
          <div
            v-for="job in proc.children"
            v-else
            :key="job.pid"
            class="flex p-2 pl-6 cursor-pointer items-center hover:bg-gray-700 hover:dark:bg-dark-gray-900"
            :class="{ 'bg-gray-700 !dark:bg-dark-gray-600': selectedProcId && selectedProcId === job.pid }"
            @click="$emit('update:selected-proc-id', job.pid)"
          >
            <div v-if="['success'].includes(job.state)" class="w-2 h-2 bg-lime-400 rounded-full" />
            <div v-if="['pending', 'skipped'].includes(job.state)" class="w-2 h-2 bg-gray-400 rounded-full" />
            <div
              v-if="['killed', 'error', 'failure', 'blocked', 'declined'].includes(job.state)"
              class="w-2 h-2 bg-red-400 rounded-full"
            />
            <div v-if="['started', 'running'].includes(job.state)" class="w-2 h-2 bg-blue-400 rounded-full" />
            <span class="ml-2">{{ job.name }}</span>
            <BuildProcDuration :proc="job" />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, toRef } from 'vue';

import BuildProcDuration from '~/components/repo/build/BuildProcDuration.vue';
import useBuild from '~/compositions/useBuild';
import { Build } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildProcList',

  components: {
    BuildProcDuration,
  },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },

    selectedProcId: {
      type: Number as PropType<number | null>,
      default: null,
    },
  },

  emits: {
    // eslint-disable-next-line @typescript-eslint/no-unused-vars
    'update:selected-proc-id': (selectedProcId: number) => true,
  },

  setup(props) {
    const build = toRef(props, 'build');
    const { prettyRef } = useBuild(build);

    return { prettyRef };
  },
});
</script>
