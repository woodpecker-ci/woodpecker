<template>
  <div v-if="build" class="flex">
    <BuildStatusIcon :build="build" class="flex items-center" />
    <div class="flex flex-col ml-4">
      <span class="underline">{{ build.owner }} / {{ build.name }}</span>
      <span>{{ message }}</span>
      <div class="flex flex-col mt-2">
        <div class="flex space-x-2 items-center">
          <icon-since />
          <span>{{ since }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <icon-duration />
          <span>{{ duration }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, ref, toRef } from 'vue';
import IconDuration from 'virtual:vite-icons/ic/sharp-timelapse';
import IconSince from 'virtual:vite-icons/mdi/clock-time-eight-outline';
import IconBranch from 'virtual:vite-icons/mdi/source-branch';
import IconCommit from 'virtual:vite-icons/mdi/source-commit';
import BuildStatusIcon from '~/components/repo/BuildStatusIcon.vue';
import { Build } from '~/lib/api/types';
import useBuild from '~/compositions/useBuild';

export default defineComponent({
  name: 'BuildFeedItem',

  components: { BuildStatusIcon, IconDuration, IconSince, IconBranch, IconCommit },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup(props) {
    const build = toRef(props, 'build');
    const { since, duration, message } = useBuild(build);

    return { build, since, duration, message };
  },
});
</script>
