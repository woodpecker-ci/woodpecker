<template>
  <IconDuration v-if="name === 'duration'" />
  <IconSince v-else-if="name === 'since'" />
  <IconBranch v-else-if="name === 'branch'" />
  <IconCommit v-else-if="name === 'commit'" />
  <IconBack v-else-if="name === 'back'" />
  <IconGithub v-else-if="name === 'github'" />
  <IconRepo v-else-if="name === 'repo'" />
  <IconSettings v-else-if="name === 'settings'" />
  <IconStatusBlocked v-else-if="name === 'status-blocked'" />
  <IconStatusDeclined v-else-if="name === 'status-declined'" />
  <IconStatusError v-else-if="name === 'status-error'" />
  <IconStatusFailure v-else-if="name === 'status-failure'" />
  <IconStatusKilled v-else-if="name === 'status-killed'" />
  <IconStatusPending v-else-if="name === 'status-pending'" />
  <IconStatusRunning v-else-if="name === 'status-running'" />
  <IconStatusSkipped v-else-if="name === 'status-skipped'" />
  <IconStatusStarted v-else-if="name === 'status-started'" />
  <IconStatusSuccess v-else-if="name === 'status-success'" />
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import IconDuration from 'virtual:vite-icons/ic/sharp-timelapse';
import IconSince from 'virtual:vite-icons/mdi/clock-time-eight-outline';
import IconBranch from 'virtual:vite-icons/mdi/source-branch';
import IconCommit from 'virtual:vite-icons/mdi/source-commit';
import IconBack from 'virtual:vite-icons/iconoir/arrow-left';
import IconGithub from 'virtual:vite-icons/mdi/github';
import IconRepo from 'virtual:vite-icons/teenyicons/git-solid';
import IconSettings from 'virtual:vite-icons/clarity/settings-solid';
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

type IconNames =
  | 'duration'
  | 'since'
  | 'branch'
  | 'commit'
  | 'back'
  | 'github'
  | 'repo'
  | 'settings'
  | 'status-blocked'
  | 'status-declined'
  | 'status-error'
  | 'status-failure'
  | 'status-killed'
  | 'status-pending'
  | 'status-running'
  | 'status-skipped'
  | 'status-started'
  | 'status-success';

export default defineComponent({
  name: 'Icon',

  components: {
    IconDuration,
    IconSince,
    IconBranch,
    IconCommit,
    IconBack,
    IconGithub,
    IconRepo,
    IconSettings,
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
    name: {
      type: String as PropType<IconNames>,
      required: true,
    },
  },
});
</script>
