<template>
  <IconDuration v-if="name === 'duration'" class="h-6 w-6" />
  <IconSince v-else-if="name === 'since'" class="h-6 w-6" />
  <IconBranch v-else-if="name === 'branch'" class="h-6 w-6" />
  <IconCommit v-else-if="name === 'commit'" class="h-6 w-6" />
  <IconBack v-else-if="name === 'back'" class="w-8 h-8" />
  <IconGithub v-else-if="name === 'github'" class="h-8 w-8" />
  <IconRepo v-else-if="name === 'repo'" class="h-8 w-8" />
  <IconSettings v-else-if="name === 'settings'" class="w-8 h-8" />
  <IconTrash v-else-if="name === 'trash'" class="h-8 w-8" />
  <IconStatusBlocked v-else-if="name === 'status-blocked'" class="h-6 w-6" />
  <IconStatusDeclined v-else-if="name === 'status-declined'" class="h-6 w-6" />
  <IconStatusError v-else-if="name === 'status-error'" class="h-8 w-8" />
  <IconStatusFailure v-else-if="name === 'status-failure'" class="h-8 w-8" />
  <IconStatusKilled v-else-if="name === 'status-killed'" class="h-7 w-7" />
  <IconStatusPending v-else-if="name === 'status-pending'" class="h-7 w-7" />
  <IconStatusRunning v-else-if="name === 'status-running'" class="h-8 w-8" />
  <IconStatusSkipped v-else-if="name === 'status-skipped'" class="h-8 w-8" />
  <IconStatusStarted v-else-if="name === 'status-started'" class="h-8 w-8" />
  <IconStatusSuccess v-else-if="name === 'status-success'" class="h-8 w-8" />
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
import IconStatusError from 'virtual:vite-icons/ph/warning';
import IconStatusFailure from 'virtual:vite-icons/ph/x-circle';
import IconStatusKilled from 'virtual:vite-icons/octicon/skip-24';
import IconStatusPending from 'virtual:vite-icons/ph/hourglass';
import IconStatusRunning from 'virtual:vite-icons/entypo/dots-two-vertical';
import IconStatusSkipped from 'virtual:vite-icons/ph/prohibit';
import IconStatusStarted from 'virtual:vite-icons/entypo/dots-two-vertical';
import IconStatusSuccess from 'virtual:vite-icons/ph/check-circle';
import IconTrash from 'virtual:vite-icons/gg/trash';

type IconNames =
  | 'duration'
  | 'since'
  | 'branch'
  | 'commit'
  | 'back'
  | 'github'
  | 'repo'
  | 'settings'
  | 'trash'
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
    IconTrash,
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
