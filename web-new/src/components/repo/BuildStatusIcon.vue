<template>
  <div v-if="build">
    <icon-status-blocked v-if="build.status === 'blocked'" class="w-6 h-6 stroke-status-red" />
    <icon-status-declined v-if="build.status === 'declined'" class="w-6 h-6 stroke-status-red" />
    <icon-status-error v-if="build.status === 'error'" class="w-6 h-6 stroke-status-red" />
    <icon-status-failure v-if="build.status === 'failure'" class="w-6 h-6 text-status-red" />
    <icon-status-killed v-if="build.status === 'killed'" class="w-6 h-6 stroke-status-red" />
    <icon-status-pending v-if="build.status === 'pending'" class="w-6 h-6 text-status-gray" />
    <icon-status-running v-if="build.status === 'running'" class="w-8 h-8 text-status-blue animate-spin animate-slow" />
    <icon-status-skipped v-if="build.status === 'skipped'" class="w-8 h-8 stroke-status-gray" />
    <icon-status-started v-if="build.status === 'started'" class="w-8 h-8 text-status-blue animate-spin animate-slow" />
    <icon-status-success v-if="build.status === 'success'" class="w-8 h-8 text-status-green" />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType, ref, toRef } from 'vue';
import IconStatusBlocked from 'virtual:vite-icons/ph/hand';
import IconStatusDeclined from 'virtual:vite-icons/ph/hand';
import IconStatusError from 'virtual:vite-icons/ph/x-circle';
import IconStatusFailure from 'virtual:vite-icons/ph/warning';
import IconStatusKilled from 'virtual:vite-icons/ph/x-circle';
import IconStatusPending from 'virtual:vite-icons/ph/hourglass';
import IconStatusRunning from 'virtual:vite-icons/entypo/dots-two-vertical';
import IconStatusSkipped from 'virtual:vite-icons/ph/prohibit';
import IconStatusStarted from 'virtual:vite-icons/entypo/dots-two-vertical';
import IconStatusSuccess from 'virtual:vite-icons/ph/check-circle';
import { Build } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildStatusIcon',

  components: {
    IconStatusBlocked,
    IconStatusDeclined,
    IconStatusError,
    IconStatusFailure,
    IconStatusKilled,
    IconStatusPending,
    IconStatusRunning,
    IconStatusSkipped,
    IconStatusStarted,
    IconStatusSuccess,
  },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },
});
</script>
