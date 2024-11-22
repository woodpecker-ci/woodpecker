<!-- cSpell:ignore radiobox timelapse -->
<template>
  <SvgIcon v-if="name === 'duration'" :path="mdiTimelapse" size="24" />
  <SvgIcon v-else-if="name === 'since'" :path="mdiClockTimeEightOutline" size="24" />
  <SvgIcon v-else-if="name === 'push'" :path="mdiSourceBranch" size="24" />
  <SvgIcon v-else-if="name === 'pull-request'" :path="mdiSourcePull" size="24" />
  <SvgIcon v-else-if="name === 'pull-request-closed'" :path="mdiSourceMerge" size="24" />
  <SvgIcon v-else-if="name === 'manual-pipeline'" :path="mdiGestureTap" size="24" />
  <SvgIcon v-else-if="name === 'tag'" :path="mdiTagOutline" size="24" />
  <SvgIcon v-else-if="name === 'deployment'" :path="mdiPackageVariant" size="24" />
  <SvgIcon v-else-if="name === 'commit'" :path="mdiSourceCommit" size="24" />
  <SvgIcon v-else-if="name === 'back'" :path="mdiArrowLeft" size="24" />
  <SvgIcon v-else-if="name === 'github'" :path="mdiGithub" size="32" />
  <SvgIcon v-else-if="name === 'repo'" :path="mdiGit" size="32" />
  <SvgIcon v-else-if="name === 'settings'" :path="mdiCog" size="32" />
  <SvgIcon v-else-if="name === 'trash'" :path="mdiTrashCanOutline" size="24" />
  <SvgIcon v-else-if="name === 'status-blocked'" :path="mdiPlayCircleOutline" size="24" />
  <SvgIcon v-else-if="name === 'status-declined'" :path="mdiStopCircleOutline" size="24" />
  <SvgIcon
    v-else-if="name === 'status-failure' || name === 'status-error' || name === 'status-killed'"
    type="mdi"
    :path="mdiCloseCircleOutline"
    size="24"
  />
  <SvgIcon v-else-if="name === 'status-pending'" :path="mdiRadioboxBlank" size="24" />
  <SvgIcon
    v-else-if="name === 'status-running' || name === 'status-started'"
    type="mdi"
    :path="mdiRadioboxIndeterminateVariant"
    size="24"
  />
  <SvgIcon v-else-if="name === 'status-skipped'" :path="mdiMinusCircleOutline" size="24" />
  <SvgIcon v-else-if="name === 'status-success'" :path="mdiCheckCircleOutline" size="24" />
  <SvgIcon v-else-if="name === 'attention'" :path="mdiAlert" size="24" />
  <SvgIcon v-else-if="name === 'warning'" :path="mdiAlertOutline" size="20" />
  <SvgIcon v-else-if="name === 'error'" :path="mdiAlertCircle" size="20" />
  <SvgIcon v-else-if="name === 'gitlab'" :path="mdiGitlab" size="32" />
  <SvgIcon v-else-if="name === 'bitbucket' || name === 'bitbucket-dc'" :path="mdiBitbucket" size="32" />
  <SvgIcon v-else-if="name === 'question'" :path="mdiHelpCircleOutline" size="24" />
  <SvgIcon v-else-if="name === 'plus'" :path="mdiPlus" size="24" />
  <SvgIcon v-else-if="name === 'list'" :path="mdiFormatListBulleted" size="24" />
  <SvgIcon v-else-if="name === 'heal'" :path="mdiBandage" size="24" />
  <SvgIcon v-else-if="name === 'turn-off'" :path="mdiPower" size="24" />
  <SvgIcon v-else-if="name === 'chevron-right'" :path="mdiChevronRight" size="24" />
  <SvgIcon v-else-if="name === 'close'" :path="mdiCloseCircleOutline" size="24" />
  <SvgIcon v-else-if="name === 'edit'" :path="mdiPencil" size="24" />
  <SvgIcon v-else-if="name === 'download'" :path="mdiDownloadCircle" size="24" />
  <SvgIcon v-else-if="name === 'stopwatch'" :path="mdiAlarm" size="24" />
  <SvgIcon v-else-if="name === 'auto-scroll'" :path="mdiDownload" size="24" />
  <SvgIcon v-else-if="name === 'auto-scroll-off'" :path="mdiDownloadOff" size="24" />
  <SvgIcon v-else-if="name === 'pause'" :path="mdiPause" size="24" />
  <SvgIcon v-else-if="name === 'play'" :path="mdiPlay" size="24" />
  <SvgIcon v-else-if="name === 'remove'" :path="mdiClose" size="24" />

  <SvgIcon v-else-if="name === 'forgejo'" :path="siForgejo.path" size="32" />
  <SvgIcon v-else-if="name === 'gitea'" :path="siGitea.path" size="32" />

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

  <div v-else-if="name === 'blank'" class="h-6 w-6" />
</template>

<script lang="ts" setup>
import {
  mdiAlarm,
  mdiAlert,
  mdiAlertCircle,
  mdiAlertOutline,
  mdiArrowLeft,
  mdiBandage,
  mdiBitbucket,
  mdiCheckCircleOutline,
  mdiChevronRight,
  mdiClockTimeEightOutline,
  mdiClose,
  mdiCloseCircleOutline,
  mdiCog,
  mdiDownload,
  mdiDownloadCircle,
  mdiDownloadOff,
  mdiFormatListBulleted,
  mdiGestureTap,
  mdiGit,
  mdiGithub,
  mdiGitlab,
  mdiHelpCircleOutline,
  mdiMinusCircleOutline,
  mdiPackageVariant,
  mdiPause,
  mdiPencil,
  mdiPlay,
  mdiPlayCircleOutline,
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
  mdiTimelapse,
  mdiTrashCanOutline,
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
  | 'pause'
  | 'warning'
  | 'attention'
  | 'spinner'
  | 'error'
  | 'remove';

defineProps<{
  name: IconNames;
}>();
</script>
