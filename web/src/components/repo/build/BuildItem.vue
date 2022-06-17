<template>
  <ListItem v-if="build" clickable class="p-0 w-full">
    <div class="flex items-center md:mr-4">
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
      <div class="w-8 flex flex-wrap justify-between items-center h-full">
        <BuildRunningIcon v-if="build.status === 'started' || build.status === 'running'" />
        <BuildStatusIcon v-else class="mx-2 md:mx-3" :build="build" />
      </div>
    </div>

    <div class="flex py-2 px-4 flex-grow min-w-0 <md:flex-wrap">
      <div class="<md:hidden flex items-center flex-shrink-0">
        <img class="w-8" :src="build.author_avatar" />
      </div>

      <div class="w-full md:w-auto md:mx-4 flex items-center min-w-0">
        <span class="text-color-alt <md:underline whitespace-nowrap overflow-hidden overflow-ellipsis">{{
          message
        }}</span>
      </div>

      <div
        class="grid grid-rows-2 grid-flow-col w-full md:ml-auto md:w-96 py-2 gap-x-4 gap-y-2 flex-shrink-0 text-color"
      >
        <div class="flex space-x-2 items-center min-w-0">
          <Icon v-if="build.event === 'pull_request'" name="pull_request" />
          <Icon v-else-if="build.event === 'deployment'" name="deployment" />
          <Icon v-else-if="build.event === 'tag'" name="tag" />
          <Icon v-else name="push" />
          <span class="truncate">{{ prettyRef }}</span>
        </div>

        <div class="flex space-x-2 items-center min-w-0">
          <Icon name="commit" />
          <span class="truncate">{{ build.commit.slice(0, 10) }}</span>
        </div>

        <div class="flex space-x-2 items-center min-w-0">
          <Icon name="duration" />
          <span class="truncate">{{ duration }}</span>
        </div>

        <div class="flex space-x-2 items-center min-w-0">
          <Icon name="since" />
          <Tooltip>
            <span>{{ since }}</span>
            <template #popper><span class="font-bold">Created</span> {{ created }}</template>
          </Tooltip>
        </div>
      </div>
    </div>
  </ListItem>
</template>

<script lang="ts">
import { Tooltip } from 'floating-vue';
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

  components: { Icon, BuildStatusIcon, ListItem, BuildRunningIcon, Tooltip },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup(props) {
    const build = toRef(props, 'build');
    const { since, duration, message, prettyRef, created } = useBuild(build);

    return { since, duration, message, prettyRef, buildStatusColors, created };
  },
});
</script>
