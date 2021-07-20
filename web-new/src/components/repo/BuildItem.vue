<template>
  <div v-if="build" class="w-full flex border rounded-md bg-white overflow-hidden hover:shadow-sm hover:bg-light-200">
    <div class="flex items-center">
      <div
        class="min-h-full w-3"
        :class="{
          'bg-gray-400': build.status === 'pending',
          'bg-green': build.status === 'started',
          'bg-green': build.status === 'success',
          'bg-red-400': build.status === 'failure',
          'bg-red-800': build.status === 'error',
          'bg-green': build.status === 'none',
        }"
      />
      <div class="w-6">
        <img v-if="build.status === 'pending'" src="../../assets/pecking_woodpecker.gif" />
      </div>
    </div>

    <div class="flex w-full py-2 px-4">
      <div class="flex items-center"><img class="w-8" :src="build.author_avatar" /></div>

      <div class="ml-4 flex items-center ml-4">
        <span>{{ build.message }}</span>
      </div>

      <div class="flex ml-auto text-gray-500">
        <div class="flex flex-col">
          <div class="space-x-2">
            <span>GH</span>
            <a class="text-gray-400" :href="build.link_url" target="_blank">{{ build.commit.slice(0, 10) }}</a>
          </div>
          <div class="space-x-2">
            <span>BR</span>
            <span>{{ build.branch }}</span>
          </div>
        </div>
        <div class="flex flex-col ml-4">
          <div class="space-x-2">
            <span>Duration</span>
            <span>{{ duration }}</span>
          </div>
          <div class="space-x-2">
            <span>Since</span>
            <span>{{ since }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, PropType, ref, toRef } from 'vue';
import { Build } from '~/lib/api/types';
import humanizeDuration from 'humanize-duration';
import TimeAgo from 'javascript-time-ago';

export default defineComponent({
  name: 'BuildItem',

  components: {},

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup(props) {
    const timeAgo = new TimeAgo('en-US');

    const build = toRef(props, 'build');

    const since = computed(() => {
      if (build.value.finished_at !== 0) {
        return timeAgo.format(Date.now() - build.value.finished_at / 1000);
      }

      if (build.value.started_at !== 0) {
        return timeAgo.format(Date.now() - build.value.started_at / 1000);
      }

      return timeAgo.format(Date.now() - build.value.enqueued_at / 1000);
    });

    const duration = computed(() => {
      if (build.value.finished_at === 0 && build.value.finished_at === 0) {
        return 'not started yet';
      }

      if (build.value.finished_at !== 0) {
        return `${humanizeDuration(Date.now() - build.value.started_at)} ...`;
      }

      return humanizeDuration(build.value.finished_at - build.value.started_at);
    });

    return { since, duration };
  },
});
</script>
