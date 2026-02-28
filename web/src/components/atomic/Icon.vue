<!-- cSpell:ignore radiobox timelapse -->
<template>
  <SvgIcon v-if="name === 'duration'" :bg-circle="bgCircle" :path="mdiTimerOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'since'" :bg-circle="bgCircle" :path="mdiClockTimeEightOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'branch'" :bg-circle="bgCircle" :path="mdiSourceBranch" size="1.3rem" />
  <SvgIcon v-else-if="name === 'pull-request'" :bg-circle="bgCircle" :path="mdiSourcePull" size="1.3rem" />
  <SvgIcon v-else-if="name === 'pull-request-closed'" :bg-circle="bgCircle" :path="mdiSourceMerge" size="1.3rem" />
  <SvgIcon v-else-if="name === 'pull-request-metadata'" :bg-circle="bgCircle" :path="mdiPencilOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'manual-pipeline'" :bg-circle="bgCircle" :path="mdiGestureTap" size="1.3rem" />
  <SvgIcon v-else-if="name === 'tag'" :bg-circle="bgCircle" :path="mdiTagOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'deployment'" :bg-circle="bgCircle" :path="mdiPackageVariant" size="1.3rem" />
  <SvgIcon v-else-if="name === 'commit'" :bg-circle="bgCircle" :path="mdiSourceCommit" size="1.3rem" />
  <SvgIcon v-else-if="name === 'back'" :bg-circle="bgCircle" :path="mdiArrowLeft" size="1.3rem" />
  <SvgIcon v-else-if="name === 'github'" :bg-circle="bgCircle" :path="mdiGithub" size="1.3rem" />
  <SvgIcon v-else-if="name === 'repo'" :bg-circle="bgCircle" :path="mdiGit" size="1.3rem" />
  <SvgIcon v-else-if="name === 'settings'" :bg-circle="bgCircle" :path="mdiCog" size="1.3rem" />
  <SvgIcon v-else-if="name === 'settings-outline'" :bg-circle="bgCircle" :path="mdiCogOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'trash'" :bg-circle="bgCircle" :path="mdiTrashCanOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'status-blocked'" :bg-circle="bgCircle" :path="mdiPlayCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'status-declined'" :bg-circle="bgCircle" :path="mdiStopCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'list-group'" :bg-circle="bgCircle" :path="mdiFormatListGroup" size="1.3rem" />
  <SvgIcon v-else-if="name === 'secret'" :bg-circle="bgCircle" :path="mdiShieldKeyOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'console'" :bg-circle="bgCircle" :path="mdiConsole" size="1.3rem" />
  <SvgIcon v-else-if="name === 'agent'" :bg-circle="bgCircle" :path="mdiPlayNetworkOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'info'" :bg-circle="bgCircle" :path="mdiInformationSlabCircleOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'user'" :bg-circle="bgCircle" :path="mdiAccountOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'org'" :bg-circle="bgCircle" :path="mdiAccountGroupOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'cron'" :bg-circle="bgCircle" :path="mdiCalendarClockOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'toolbox'" :bg-circle="bgCircle" :path="mdiToolboxOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'puzzle'" :bg-circle="bgCircle" :path="mdiPuzzleOutline" size="1.3rem" />
  <SvgIcon
    v-else-if="name === 'status-failure' || name === 'status-error' || name === 'status-killed'"
    type="mdi"
    :bg-circle="bgCircle"
    :path="mdiCloseCircle"
    size="1.3rem"
  />
  <SvgIcon
    v-else-if="name === 'status-pending' || name === 'status-created'"
    :bg-circle="bgCircle"
    :path="mdiRadioboxBlank"
    size="1.3rem"
  />
  <SvgIcon
    v-else-if="name === 'status-running' || name === 'status-started'"
    type="mdi"
    :bg-circle="bgCircle"
    :path="mdiRadioboxIndeterminateVariant"
    size="1.3rem"
  />
  <SvgIcon v-else-if="name === 'status-skipped'" :bg-circle="bgCircle" :path="mdiMinusCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'status-success'" :bg-circle="bgCircle" :path="mdiCheckCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'alert'" :bg-circle="bgCircle" :path="mdiAlertCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'question'" :bg-circle="bgCircle" :path="mdiHelpCircle" size="1.3rem" />
  <SvgIcon v-else-if="name === 'plus'" :bg-circle="bgCircle" :path="mdiPlus" size="1.3rem" />
  <SvgIcon v-else-if="name === 'refresh'" :bg-circle="bgCircle"  :path="mdiRefresh" size="1.3rem" />
  <SvgIcon v-else-if="name === 'list'" :bg-circle="bgCircle" :path="mdiFormatListBulleted" size="1.3rem" />
  <SvgIcon v-else-if="name === 'heal'" :bg-circle="bgCircle" :path="mdiWrenchCogOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'turn-off'" :bg-circle="bgCircle" :path="mdiPower" size="1.3rem" />
  <SvgIcon v-else-if="name === 'chevron-right'" :bg-circle="bgCircle" :path="mdiChevronRight" size="1.3rem" />
  <SvgIcon v-else-if="name === 'close'" :bg-circle="bgCircle" :path="mdiClose" size="1.3rem" />
  <SvgIcon v-else-if="name === 'edit'" :bg-circle="bgCircle" :path="mdiPencilOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'download'" :bg-circle="bgCircle" :path="mdiDownloadOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'stopwatch'" :bg-circle="bgCircle" :path="mdiAlarm" size="1.3rem" />
  <SvgIcon v-else-if="name === 'auto-scroll'" :bg-circle="bgCircle" :path="mdiEyeOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'auto-scroll-off'" :bg-circle="bgCircle" :path="mdiEyeOffOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'pause'" :bg-circle="bgCircle" :path="mdiPause" size="1.3rem" />
  <SvgIcon v-else-if="name === 'play'" :bg-circle="bgCircle" :path="mdiPlay" size="1.3rem" />
  <SvgIcon v-else-if="name === 'play-outline'" :bg-circle="bgCircle" :path="mdiPlayOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'dots'" :bg-circle="bgCircle" :path="mdiDotsVertical" size="1.3rem" />
  <SvgIcon v-else-if="name === 'tray-full'" :bg-circle="bgCircle" :path="mdiTrayFull" size="1.3rem" />
  <SvgIcon v-else-if="name === 'file-cog-outline'" :bg-circle="bgCircle" :path="mdiFileCogOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'file-edit-outline'" :bg-circle="bgCircle" :path="mdiFileEditOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'folder'" :bg-circle="bgCircle" :path="mdiFolderOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'folder-open'" :bg-circle="bgCircle" :path="mdiFolderOpenOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'file'" :bg-circle="bgCircle" :path="mdiFileOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'bug-outline'" :bg-circle="bgCircle" :path="mdiBugOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'docker'" :bg-circle="bgCircle" :path="mdiDocker" size="1.3rem" />
  <SvgIcon v-else-if="name === 'forge'" :bg-circle="bgCircle" :path="mdiCodeBraces" size="1.3rem" />
  <SvgIcon v-else-if="name === 'fullscreen'" :bg-circle="bgCircle" :path="mdiFullscreen" size="1.3rem" />
  <SvgIcon v-else-if="name === 'exit-fullscreen'" :bg-circle="bgCircle" :path="mdiFullscreenExit" size="1.3rem" />

  <SvgIcon v-else-if="name === 'visibility-private'" :bg-circle="bgCircle" :path="mdiLockOutline" size="1.3rem" />
  <SvgIcon v-else-if="name === 'visibility-internal'" :bg-circle="bgCircle" :path="mdiLockOpenOutline" size="1.3rem" />

  <SvgIcon v-else-if="name === 'forgejo'" :bg-circle="bgCircle" :path="siForgejo.path" size="1.3rem" />
  <SvgIcon v-else-if="name === 'gitea'" :bg-circle="bgCircle" :path="siGitea.path" size="1.3rem" />
  <SvgIcon v-else-if="name === 'gitlab'" :bg-circle="bgCircle" :path="mdiGitlab" size="1.3rem" />
  <SvgIcon
    v-else-if="name === 'bitbucket' || name === 'bitbucket-dc'"
    :bg-circle="bgCircle"
    :path="mdiBitbucket"
    size="1.3rem"
  />

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
  mdiCodeBraces,
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
  mdiFileOutline,
  mdiFolderOpenOutline,
  mdiFolderOutline,
  mdiFormatListBulleted,
  mdiFormatListGroup,
  mdiFullscreen,
  mdiFullscreenExit,
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
  mdiPuzzleOutline,
  mdiRadioboxBlank,
  mdiRadioboxIndeterminateVariant,
  mdiRefresh,
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
} from '@mdi/js';
import { siForgejo, siGitea } from 'simple-icons';

import SvgIcon from './SvgIcon.vue';

export type IconNames =
  | 'duration'
  | 'since'
  | 'branch'
  | 'pull-request'
  | 'pull-request-closed'
  | 'pull-request-metadata'
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
  | 'status-created'
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
  | 'refresh'
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
  | 'puzzle'
  | 'forge'
  | 'fullscreen'
  | 'exit-fullscreen'
  | 'folder'
  | 'folder-open'
  | 'file';

const props = defineProps<{
  name: IconNames;
  bgCircle?: boolean;
}>();

function throwNotFound() {
  throw new Error(`Icon "${props.name}" not found`);
}
</script>
