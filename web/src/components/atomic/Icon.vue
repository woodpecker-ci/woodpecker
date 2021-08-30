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

import IconDuration from '~icons/ic/sharp-timelapse';
import IconSince from '~icons/mdi/clock-time-eight-outline';
import IconBranch from '~icons/mdi/source-branch';
import IconCommit from '~icons/mdi/source-commit';
import IconBack from '~icons/iconoir/arrow-left';
import IconGithub from '~icons/mdi/github';
import IconRepo from '~icons/teenyicons/git-solid';
import IconSettings from '~icons/clarity/settings-solid';
import IconStatusBlocked from '~icons/ph/hand';
import IconStatusDeclined from '~icons/ph/hand';
import IconStatusError from '~icons/ph/warning';
import IconStatusFailure from '~icons/ph/x-circle';
import IconStatusKilled from '~icons/octicon/skip-24';
import IconStatusPending from '~icons/ph/hourglass';
import IconStatusRunning from '~icons/entypo/dots-two-vertical';
import IconStatusSkipped from '~icons/ph/prohibit';
import IconStatusStarted from '~icons/entypo/dots-two-vertical';
import IconStatusSuccess from '~icons/ph/check-circle';
import IconTrash from '~icons/gg/trash';

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
