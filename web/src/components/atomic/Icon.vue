<!-- cSpell:ignore radiobox timelapse -->
<template>
  <SvgIcon v-if="name === 'duration'" :path="mdiTimerOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'since'" :path="mdiClockTimeEightOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'branch'" :path="mdiSourceBranch" size="1.3rem" />
  <SvgIcon v-else-if="name === 'pull-request'" :path="mdiSourcePull" size="1.3rem" />
  <SvgIcon v-else-if="name === 'pull-request-closed'" :path="mdiSourceMerge" size="1.3rem" />
  <SvgIcon v-else-if="name === 'manual-pipeline'" :path="mdiGestureTap" size="1.3rem" />
  <SvgIcon v-else-if="name === 'tag'" :path="mdiTagOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'deployment'" :path="mdiPackageVariant" size="1.3rem" />
  <SvgIcon v-else-if="name === 'commit'" :path="mdiSourceCommit" size="1.3rem" />
  <SvgIcon v-else-if="name === 'back'" :path="mdiArrowLeft" size="1.3rem" />
  <SvgIcon v-else-if="name === 'github'" :path="mdiGithub" size="1.3rem" />
  <SvgIcon v-else-if="name === 'repo'" :path="mdiGit" size="1.3rem" />
  <SvgIcon v-else-if="name === 'settings'" :path="mdiCog" size="1.3rem" />
  <SvgIcon v-else-if="name === 'settings-outline'" :path="mdiCogOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'trash'" :path="mdiTrashCanOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'status-blocked'" :path="mdiPlayCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'status-declined'" :path="mdiStopCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'list-group'" :path="mdiFormatListGroup" size="1.3rem" />
  <SvgIcon v-else-if="name === 'secret'" :path="mdiShieldKeyOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'console'" :path="mdiConsole" size="1.3rem" />
  <SvgIcon v-else-if="name === 'agent'" :path="mdiPlayNetworkOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'info'" :path="mdiInformationSlabCircleOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'user'" :path="mdiAccountOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'org'" :path="mdiAccountGroupOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'cron'" :path="mdiCalendarClockOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'toolbox'" :path="mdiToolboxOutline" size="1.3rem" />
  <SvgIcon
    v-else-if="name === 'status-failure' || name === 'status-error' || name === 'status-killed'"
    type="mdi"
    :path="mdiCloseCircle"
    size="1.3rem"
  />
  <SvgIcon v-else-if="name === 'status-pending'" :path="mdiRadioboxBlank" size="1.3rem" />
  <SvgIcon
    v-else-if="name === 'status-running' || name === 'status-started'"
    type="mdi"
    :path="mdiRadioboxIndeterminateVariant"
    size="1.3rem"
  />
  <SvgIcon v-else-if="name === 'status-skipped'" :path="mdiMinusCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'status-success'" :path="mdiCheckCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'alert'" :path="mdiAlertCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'question'" :path="mdiHelpCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'plus'" :path="mdiPlus" size="1.3rem" />
  <SvgIcon v-else-if="name === 'list'" :path="mdiFormatListBulleted" size="1.3rem" />
  <SvgIcon v-else-if="name === 'heal'" :path="mdiWrenchCogOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'turn-off'" :path="mdiPower" size="1.3rem" />
  <SvgIcon v-else-if="name === 'chevron-right'" :path="mdiChevronRight" size="1.3rem" />
  <SvgIcon v-else-if="name === 'close'" :path="mdiClose" size="1.3rem" />
  <SvgIcon v-else-if="name === 'edit'" :path="mdiPencilOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'download'" :path="mdiDownloadOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'stopwatch'" :path="mdiAlarm" size="1.3rem" />
  <SvgIcon v-else-if="name === 'auto-scroll'" :path="mdiEyeOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'auto-scroll-off'" :path="mdiEyeOffOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'pause'" :path="mdiPause" size="1.3rem" />
  <SvgIcon v-else-if="name === 'play'" :path="mdiPlay" size="1.3rem" />
  <SvgIcon v-else-if="name === 'play-outline'" :path="mdiPlayOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'dots'" :path="mdiDotsVertical" size="1.3rem" />
  <SvgIcon v-else-if="name === 'tray-full'" :path="mdiTrayFull" size="1.3rem" />
  <SvgIcon v-else-if="name === 'file-cog-outline'" :path="mdiFileCogOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'file-edit-outline'" :path="mdiFileEditOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'bug-outline'" :path="mdiBugOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'docker'" :path="mdiDocker" size="1.3rem" />
  <SvgIcon v-else-if="name === 'forge'" :path="mdiCodeBraces" size="1.3rem" />

  <SvgIcon v-else-if="name === 'visibility-private'" :path="mdiLockOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'visibility-internal'" :path="mdiLockOpenOutline" size="1.3rem" />

  <SvgIcon v-else-if="name === 'forgejo'" :path="siForgejo.path" size="1.3rem" />
  <SvgIcon v-else-if="name === 'gitea'" :path="siGitea.path" size="1.3rem" />
  <SvgIcon v-else-if="name === 'gitlab'" :path="mdiGitlab" size="1.3rem" />
  <SvgIcon v-else-if="name === 'bitbucket' || name === 'bitbucket-dc'" :path="mdiBitbucket" size="1.3rem" />

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

  <div v-else>{{ throwNotFound() }}</div>
</template>

<script lang="ts" setup>
import {
  mdiAccountGroupOutline,
  mdiAccountOutline,
  mdiAlarm,
  mdiAlertCircle,
  mdiArrowLeft,
  mdiBitbucket,
  mdiBugOutline,
  mdiCalendarClockOutline,
  mdiCheckCircle,
  mdiChevronRight,
  mdiClockTimeEightOutline,
  mdiClose,
  mdiCloseCircle,
  mdiCog,
  mdiCogOutline,
  mdiConsole,
  mdiDocker,
  mdiDotsVertical,
  mdiDownloadOutline,
  mdiEyeOffOutline,
  mdiEyeOutline,
  mdiFileCogOutline,
  mdiFileEditOutline,
  mdiFormatListBulleted,
  mdiFormatListGroup,
  mdiGestureTap,
  mdiGit,
  mdiGithub,
  mdiGitlab,
  mdiHelpCircle,
  mdiInformationSlabCircleOutline,
  mdiLockOpenOutline,
  mdiLockOutline,
  mdiMinusCircle,
  mdiPackageVariant,
  mdiPause,
  mdiPencilOutline,
  mdiPlay,
  mdiPlayCircle,
  mdiPlayNetworkOutline,
  mdiPlayOutline,
  mdiPlus,
  mdiPower,
  mdiRadioboxBlank,
  mdiRadioboxIndeterminateVariant,
  mdiShieldKeyOutline,
  mdiSourceBranch,
  mdiSourceCommit,
  mdiSourceMerge,
  mdiSourcePull,
  mdiStopCircle,
  mdiTagOutline,
  mdiTimerOutline,
  mdiToolboxOutline,
  mdiTrashCanOutline,
  mdiTrayFull,
  mdiWrenchCogOutline,
  mdiCodeBraces,
} from '@mdi/js';
import { siForgejo, siGitea } from 'simple-icons';

import SvgIcon from './SvgIcon.vue';

export type IconNames =
  | 'duration'
  | 'since'
  | 'branch'
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
  | 'play'
  | 'play-outline'
  | 'pause'
  | 'alert'
  | 'spinner'
  | 'visibility-private'
  | 'visibility-internal'
  | 'dots'
  | 'tray-full'
  | 'file-cog-outline'
  | 'file-edit-outline'
  | 'bug-outline'
  | 'list-group'
  | 'secret'
  | 'docker'
  | 'console'
  | 'agent'
  | 'info'
  | 'repo'
  | 'user'
  | 'org'
  | 'cron'
  | 'toolbox'
  | 'forge';

const props = defineProps<{
  name: IconNames;
}>();

function throwNotFound() {
  throw new Error(`Icon "${props.name}" not found`);
}
</script>
