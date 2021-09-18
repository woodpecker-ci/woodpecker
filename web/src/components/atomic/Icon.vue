<template>
  <IconDuration v-if="name === 'duration'" class="h-6 w-6" />
  <IconSince v-else-if="name === 'since'" class="h-6 w-6" />
  <IconPush v-else-if="name === 'push'" class="h-6 w-6" />
  <IconPullRequest v-else-if="name === 'pull_request'" class="h-6 w-6" />
  <IconTag v-else-if="name === 'tag'" class="h-6 w-6" />
  <IconDeployment v-else-if="name === 'deployment'" class="h-6 w-6" />
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
  <IconGitea v-else-if="name === 'gitea'" class="h-8 w-8" />
  <IconQuestion v-else-if="name === 'question'" class="h-6 w-6" />
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import IconGitea from '~icons/cib/gitea';
import IconDeployment from '~icons/clarity/deploy-line';
import IconSettings from '~icons/clarity/settings-solid';
import IconStatusRunning from '~icons/entypo/dots-two-vertical';
import IconStatusStarted from '~icons/entypo/dots-two-vertical';
import IconTrash from '~icons/gg/trash';
import IconDuration from '~icons/ic/sharp-timelapse';
import IconBack from '~icons/iconoir/arrow-left';
import IconSince from '~icons/mdi/clock-time-eight-outline';
import IconGithub from '~icons/mdi/github';
import IconPush from '~icons/mdi/source-branch';
import IconCommit from '~icons/mdi/source-commit';
import IconPullRequest from '~icons/mdi/source-pull';
import IconTag from '~icons/mdi/tag-outline';
import IconStatusKilled from '~icons/octicon/skip-24';
import IconStatusSuccess from '~icons/ph/check-circle';
import IconStatusBlocked from '~icons/ph/hand';
import IconStatusDeclined from '~icons/ph/hand';
import IconStatusPending from '~icons/ph/hourglass';
import IconStatusSkipped from '~icons/ph/prohibit';
import IconStatusError from '~icons/ph/warning';
import IconStatusFailure from '~icons/ph/x-circle';
import IconRepo from '~icons/teenyicons/git-solid';
import IconQuestion from '~icons/vaadin/question-circle-o';

type IconNames =
  | 'duration'
  | 'since'
  | 'push'
  | 'pull_request'
  | 'tag'
  | 'deployment'
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
  | 'status-success'
  | 'gitea'
  | 'question';

export default defineComponent({
  name: 'Icon',

  components: {
    IconDuration,
    IconSince,
    IconPush,
    IconPullRequest,
    IconDeployment,
    IconTag,
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
    IconGitea,
    IconQuestion,
  },

  props: {
    name: {
      type: String as PropType<IconNames>,
      required: true,
    },
  },
});
</script>
