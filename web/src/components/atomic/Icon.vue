<!-- cSpell:ignore radiobox timelapse -->
<template>
  <SvgIcon v-if="name === 'duration'" :path="mdiTimerOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'since'" :path="mdiClockTimeEightOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'push'" :path="mdiSourceBranch" size="1.2rem" />
  <SvgIcon v-else-if="name === 'pull-request'" :path="mdiSourcePull" size="1.2rem" />
  <SvgIcon v-else-if="name === 'pull-request-closed'" :path="mdiSourceMerge" size="1.2rem" />
  <SvgIcon v-else-if="name === 'manual-pipeline'" :path="mdiGestureTap" size="1.2rem" />
  <SvgIcon v-else-if="name === 'tag'" :path="mdiTagOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'deployment'" :path="mdiPackageVariant" size="1.2rem" />
  <SvgIcon v-else-if="name === 'commit'" :path="mdiSourceCommit" size="1.2rem" />
  <SvgIcon v-else-if="name === 'back'" :path="mdiArrowLeft" size="1.2rem" />
  <SvgIcon v-else-if="name === 'github'" :path="mdiGithub" size="1.2rem" />
  <SvgIcon v-else-if="name === 'repo'" :path="mdiGit" size="1.2rem" />
  <SvgIcon v-else-if="name === 'settings'" :path="mdiCog" size="1.2rem" />
  <SvgIcon v-else-if="name === 'settings-outline'" :path="mdiCogOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'trash'" :path="mdiTrashCanOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'status-blocked'" :path="mdiPlayOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'status-declined'" :path="mdiStopCircleOutline" size="1.2rem" />
  <SvgIcon
    v-else-if="name === 'status-failure' || name === 'status-error' || name === 'status-killed'"
    type="mdi"
    :path="mdiCloseCircle"
    size="1.2rem"
  />
  <SvgIcon v-else-if="name === 'status-pending'" :path="mdiRadioboxBlank" size="1.2rem" />
  <SvgIcon
    v-else-if="name === 'status-running' || name === 'status-started'"
    type="mdi"
    :path="mdiRadioboxIndeterminateVariant"
    size="1.2rem"
  />
  <SvgIcon v-else-if="name === 'status-skipped'" :path="mdiMinusCircle" size="1.2rem" />
  <SvgIcon v-else-if="name === 'status-success'" :path="mdiCheckCircle" size="1.2rem" />
  <SvgIcon v-else-if="name === 'attention'" :path="mdiAlertCircle" size="1.2rem" />
  <SvgIcon v-else-if="name === 'warning'" :path="mdiAlert" size="1.2rem" />
  <SvgIcon v-else-if="name === 'error'" :path="mdiAlert" size="1.2rem" />
  <SvgIcon v-else-if="name === 'bitbucket' || name === 'bitbucket-dc'" :path="mdiBitbucket" size="32" />
  <SvgIcon v-else-if="name === 'question'" :path="mdiHelpCircle" size="1.2rem" />
  <SvgIcon v-else-if="name === 'plus'" :path="mdiPlus" size="1.2rem" />
  <SvgIcon v-else-if="name === 'list'" :path="mdiFormatListBulleted" size="1.2rem" />
  <SvgIcon v-else-if="name === 'heal'" :path="mdiWrenchCogOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'turn-off'" :path="mdiPower" size="1.2rem" />
  <SvgIcon v-else-if="name === 'chevron-right'" :path="mdiChevronRight" size="1.2rem" />
  <SvgIcon v-else-if="name === 'close'" :path="mdiCloseCircle" size="1.2rem" />
  <SvgIcon v-else-if="name === 'edit'" :path="mdiPencilOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'download'" :path="mdiDownloadCircle" size="1.2rem" />
  <SvgIcon v-else-if="name === 'stopwatch'" :path="mdiAlarm" size="1.2rem" />
  <SvgIcon v-else-if="name === 'auto-scroll'" :path="mdiDownloadCircle" size="1.2rem" />
  <SvgIcon v-else-if="name === 'auto-scroll-off'" :path="mdiDownloadOff" size="1.2rem" />
  <SvgIcon v-else-if="name === 'pause'" :path="mdiPause" size="1.2rem" />
  <SvgIcon v-else-if="name === 'play'" :path="mdiPlay" size="1.2rem" />
  <SvgIcon v-else-if="name === 'play-outline'" :path="mdiPlayOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'remove'" :path="mdiCloseCircle" size="1.2rem" />

  <SvgIcon v-else-if="name === 'visibility-private'" :path="mdiLockOutline" size="1.2rem" />
  <SvgIcon v-else-if="name === 'visibility-internal'" :path="mdiLockOpenOutline" size="1.2rem" />

  <SvgIcon v-else-if="name === 'forgejo'" :path="siForgejo.path" size="1.2rem" />
  <SvgIcon v-else-if="name === 'gitea'" :path="siGitea.path" size="1.2rem" />

  <svg v-else-if="name === 'spinner'" width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
    <path
      fill="currentColor"
      d="M12,1A11,11,0,1,0,23,12,11,11,0,0,0,12,1Zm0,19a8,8,0,1,1,8-8A8,8,0,0,1,12,20Z"
      opacity=".25"
    />
    <path
      fill="currentColor"
      d="M12,4a8,8,0,0,1,7.89,6.7A1.53,1.53,0,0,0,21.38,12h0a1.5,1.5,0,0,0,1.48-1.75,11,11,0,0,0-21.72,0A1.5,1.5,0,0,0,2.62,12h0a1.53,1.53,0,0,0,1.49-1.3A8,8,0,0,1,12,4Z"
    >
      <animateTransform
        attributeName="transform"
        type="rotate"
        dur="0.75s"
        values="0 12 12;360 12 12"
        repeatCount="indefinite"
      />
    </path>
  </svg>

  <div v-else-if="name === 'blank'" class="w-6 h-6" />
</template>

<script lang="ts" setup>
import {
  mdiAlarm,
  mdiAlert,
  mdiAlertCircle,
  mdiArrowLeft,
  mdiBitbucket,
  mdiCheckCircle,
  mdiChevronRight,
  mdiClockTimeEightOutline,
  mdiCloseCircle,
  mdiCog,
  mdiCogOutline,
  mdiDownloadCircle,
  mdiDownloadOff,
  mdiFormatListBulleted,
  mdiGestureTap,
  mdiGit,
  mdiGithub,
  mdiHelpCircle,
  mdiLockOpenOutline,
  mdiLockOutline,
  mdiMinusCircle,
  mdiPackageVariant,
  mdiPause,
  mdiPencilOutline,
  mdiPlay,
  mdiPlayOutline,
  mdiPlus,
  mdiPower,
  mdiRadioboxBlank,
  mdiRadioboxIndeterminateVariant,
  mdiSourceBranch,
  mdiSourceCommit,
  mdiSourceMerge,
  mdiSourcePull,
  mdiStopCircleOutline,
  mdiTagOutline,
  mdiTimerOutline,
  mdiTrashCanOutline,
  mdiWrenchCogOutline,
} from '@mdi/js';
import { siForgejo, siGitea } from 'simple-icons';

import SvgIcon from './SvgIcon.vue';

export type IconNames =
  | 'duration'
  | 'since'
  | 'push'
  | 'pull-request'
  | 'pull-request-closed'
  | 'manual-pipeline'
  | 'tag'
  | 'deployment'
  | 'commit'
  | 'back'
  | 'github'
  | 'repo'
  | 'settings'
  | 'settings-outline'
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
  | 'gitlab'
  | 'bitbucket'
  | 'bitbucket-dc'
  | 'forgejo'
  | 'question'
  | 'list'
  | 'loading'
  | 'plus'
  | 'blank'
  | 'heal'
  | 'chevron-right'
  | 'turn-off'
  | 'close'
  | 'edit'
  | 'stopwatch'
  | 'download'
  | 'auto-scroll'
  | 'auto-scroll-off'
  | 'refresh'
  | 'play'
  | 'play-outline'
  | 'pause'
  | 'warning'
  | 'attention'
  | 'spinner'
  | 'error'
  | 'remove'
  | 'visibility-private'
  | 'visibility-internal';

defineProps<{
  name: IconNames;
}>();
</script>
