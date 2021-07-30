<template>
  <div v-if="build" class="w-full flex border rounded-md bg-white overflow-hidden hover:shadow-sm hover:bg-light-200">
    <div class="flex items-center mr-4">
      <div
        class="min-h-full w-3"
        :class="{
          'bg-status-red': ['killed', 'error', 'failure', 'blocked', 'declined'].includes(build.status),
          'bg-status-gray': ['started', 'running'].includes(build.status),
          'bg-status-green': ['success'].includes(build.status),
          'bg-status-blue': ['started', 'running'].includes(build.status),
        }"
      />
      <div class="w-8">
        <img
          v-if="build.status === 'started' || build.status === 'running'"
          class="w-6"
          src="../../assets/pecking_woodpecker.gif"
        />
        <BuildStatusIcon class="mx-3" v-else :build="build" />
      </div>
    </div>

    <div class="flex w-full py-2 px-4">
      <div class="flex items-center"><img class="w-8" :src="build.author_avatar" /></div>

      <div class="ml-4 flex items-center ml-4">
        <span>{{ message }}</span>
      </div>

      <div class="flex ml-auto text-gray-500 py-2">
        <div class="flex flex-col space-y-2 w-42">
          <div class="flex space-x-2 items-center">
            <icon-commit />
            <a class="text-gray-400" :href="build.link_url" target="_blank">{{ build.commit.slice(0, 10) }}</a>
          </div>
          <div class="flex space-x-2 items-center">
            <icon-branch />
            <span>{{ build.branch }}</span>
          </div>
        </div>
        <div class="flex flex-col ml-4 space-y-2 w-42">
          <div class="flex space-x-2 items-center">
            <icon-duration />
            <span>{{ duration }}</span>
          </div>
          <div class="flex space-x-2 items-center">
            <icon-since />
            <span>{{ since }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, ref, toRef } from 'vue';
import IconDuration from 'virtual:vite-icons/ic/sharp-timelapse.vue';
import IconSince from 'virtual:vite-icons/mdi/clock-time-eight-outline.vue';
import IconBranch from 'virtual:vite-icons/mdi/source-branch.vue';
import IconCommit from 'virtual:vite-icons/mdi/source-commit.vue';
import { Build } from '~/lib/api/types';
import useBuild from '~/compositions/useBuild';
import BuildStatusIcon from './BuildStatusIcon.vue';

export default defineComponent({
  name: 'BuildItem',

  components: { IconDuration, IconSince, IconBranch, IconCommit, BuildStatusIcon },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup(props) {
    const build = toRef(props, 'build');
    const { since, duration, message } = useBuild(build);

    return { since, duration, message };
  },
});
</script>
