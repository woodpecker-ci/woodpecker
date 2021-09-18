<template>
  <ListItem v-if="build" clickable class="p-0">
    <div class="flex items-center mr-4">
      <div
        class="min-h-full w-3"
        :class="{
          'bg-yellow-400': build.status === 'pending',
          'bg-status-red': buildStatusColors[build.status] === 'red',
          'bg-status-gray': buildStatusColors[build.status] === 'gray',
          'bg-status-green': buildStatusColors[build.status] === 'green',
          'bg-status-blue': buildStatusColors[build.status] === 'blue',
        }"
      />
      <div class="w-8 flex">
        <img
          v-if="build.status === 'started' || build.status === 'running'"
          class="w-6"
          src="../../assets/pecking_woodpecker.gif"
        />
        <BuildStatusIcon v-else class="mx-3" :build="build" />
      </div>
    </div>

    <div class="flex w-full py-2 px-4">
      <div class="flex items-center"><img class="w-8" :src="build.author_avatar" /></div>

      <div class="ml-4 flex items-center mx-4">
        <span>{{ message }}</span>
      </div>

      <div class="flex ml-auto text-gray-500 py-2">
        <div class="flex flex-col space-y-2 w-42">
          <div class="flex space-x-2 items-center">
            <Icon v-if="build.event === 'pull_request'" name="pull_request" />
            <Icon v-else-if="build.event === 'deployment'" name="deployment" />
            <Icon v-else-if="build.event === 'tag'" name="tag" />
            <Icon v-else name="push" />
            <span class="truncate">{{ build.branch }}</span>
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
import BuildStatusIcon from '~/components/repo/BuildStatusIcon.vue';
import useBuild from '~/compositions/useBuild';
import { Build } from '~/lib/api/types';

import { buildStatusColors } from './build-status';

export default defineComponent({
  name: 'BuildItem',

  components: { Icon, BuildStatusIcon, ListItem },

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
