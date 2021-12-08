<template>
  <ListItem v-if="build" clickable class="p-0 w-full">
    <div class="flex items-center mr-4">
      <div
        class="min-h-full w-3"
        :class="{
          'bg-yellow-400 dark:bg-dark-200': build.status === 'pending',
          'bg-red-400 dark:bg-red-800': buildStatusColors[build.status] === 'red',
          'bg-gray-600 dark:bg-gray-500': buildStatusColors[build.status] === 'gray',
          'bg-lime-400 dark:bg-lime-900': buildStatusColors[build.status] === 'green',
          'bg-blue-400 dark:bg-blue-900': buildStatusColors[build.status] === 'blue',
        }"
      />
      <div class="w-8 flex">
        <BuildRunningIcon v-if="build.status === 'started' || build.status === 'running'" />
        <BuildStatusIcon v-else class="mx-3" :build="build" />
      </div>
    </div>

    <div class="flex py-2 px-4 flex-grow min-w-0">
      <div class="flex items-center flex-shrink-0"><img class="w-8" :src="build.author_avatar" /></div>

      <div class="ml-4 flex items-center mx-4 min-w-0">
        <span class="text-gray-600 dark:text-gray-500 whitespace-nowrap overflow-hidden overflow-ellipsis">{{
          message
        }}</span>
      </div>

      <div class="flex ml-auto text-gray-500 py-2 flex-shrink-0">
        <div class="flex flex-col space-y-2 w-42">
          <div class="flex space-x-2 items-center">
            <Icon v-if="build.event === 'pull_request'" name="pull_request" />
            <Icon v-else-if="build.event === 'deployment'" name="deployment" />
            <Icon v-else-if="build.event === 'tag'" name="tag" />
            <Icon v-else name="push" />
            <span v-if="build.event === 'pull_request'" class="truncate">{{
              `#${build.ref.replaceAll('refs/pull/', '').replaceAll('/merge', '').replaceAll('/head', '')}`
            }}</span>
            <span v-else class="truncate">{{ build.branch }}</span>
          </div>
          <div class="flex space-x-2 items-center">
            <Icon name="commit" />
            <span>{{ build.commit.slice(0, 10) }}</span>
          </div>
        </div>
        <div class="flex flex-col ml-4 space-y-2 w-42">
          <div class="flex space-x-2 items-center">
            <Icon name="duration" />
            <span>{{ duration }}</span>
          </div>
          <div class="flex space-x-2 items-center">
            <Icon name="since" />
            <span>{{ since }}</span>
          </div>
        </div>
      </div>
    </div>
  </ListItem>
</template>

<script lang="ts">
import { defineComponent, PropType, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import { buildStatusColors } from '~/components/repo/build/build-status';
import BuildRunningIcon from '~/components/repo/build/BuildRunningIcon.vue';
import BuildStatusIcon from '~/components/repo/build/BuildStatusIcon.vue';
import useBuild from '~/compositions/useBuild';
import { Build } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildItem',

  components: { Icon, BuildStatusIcon, ListItem, BuildRunningIcon },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup(props) {
    const build = toRef(props, 'build');
    const { since, duration, message } = useBuild(build);

    return { since, duration, message, buildStatusColors };
  },
});
</script>
