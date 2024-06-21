<!-- cSpell:ignore radiobox timelapse -->
<template>
  <SvgIcon type="mdi" :path="mdiTimelapse" v-if="name === 'duration'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiClockTimeEightOutline" v-else-if="name === 'since'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiSourceBranch" v-else-if="name === 'push'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiSourcePull" v-else-if="name === 'pull-request'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiSourceMerge" v-else-if="name === 'pull-request-closed'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiGestureTap" v-else-if="name === 'manual-pipeline'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiTagOutline" v-else-if="name === 'tag'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiPackageVariant" v-else-if="name === 'deployment'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiSourceCommit" v-else-if="name === 'commit'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiArrowLeft" v-else-if="name === 'back'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiGithub" v-else-if="name === 'github'" :size="32"/>
  <SvgIcon type="mdi" :path="mdiGit" v-else-if="name === 'repo'" :size="32"/>
  <SvgIcon type="mdi" :path="mdiCog" v-else-if="name === 'settings'" :size="32"/>
  <SvgIcon type="mdi" :path="mdiTrashCanOutline" v-else-if="name === 'trash'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiPlay" v-else-if="name === 'status-blocked' || name === 'play'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiStop" v-else-if="name === 'status-declined'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiCloseThick" v-else-if="name === 'status-failure' || name === 'status-error' || name === 'status-killed'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiRadioboxBlank" v-else-if="name === 'status-pending'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiRadioboxIndeterminateVariant" v-else-if="name === 'status-running' || name === 'status-started'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiMinusCircle" v-else-if="name === 'status-skipped'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiCheckCircle" v-else-if="name === 'status-success'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiAlert" v-else-if="name === 'attention'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiAlertOutline" v-else-if="name === 'warning'" :size="20"/>
  <SvgIcon type="mdi" :path="mdiAlertCircle" v-else-if="name === 'error'" :size="20"/>
  <SvgIcon type="mdi" :path="mdiGitlab" v-else-if="name === 'gitlab'" :size="32"/>
  <SvgIcon type="mdi" :path="mdiBitbucket" v-else-if="name === 'bitbucket' || name === 'bitbucket-dc'" :size="32"/>
  <SvgIcon type="mdi" :path="mdiHelpCircleOutline" v-else-if="name === 'question'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiPlus" v-else-if="name === 'plus'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiFormatListBulleted" v-else-if="name === 'list'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiBandage" v-else-if="name === 'heal'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiPower" v-else-if="name === 'turn-off'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiChevronRight" v-else-if="name === 'chevron-right'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiCloseCircleOutline" v-else-if="name === 'close'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiPencil" v-else-if="name === 'edit'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiDownloadCircle" v-else-if="name === 'download'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiAlarm" v-else-if="name === 'stopwatch'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiDownload" v-else-if="name === 'auto-scroll'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiDownloadOff" v-else-if="name === 'auto-scroll-off'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiPause" v-else-if="name === 'pause'" :size="24"/>
  <SvgIcon type="mdi" :path="mdiClose" v-else-if="name === 'times'" :size="24"/>

  <SvgIcon type="simple-icons" :path="siForgejo.path" v-else-if="name === 'forgejo'" class="h-8 w-8"/>
  <SvgIcon type="simple-icons" :path="siGitea.path" v-else-if="name === 'gitea'" :size="32"/>

  <svg v-else-if="name === 'spinner'" width="24" height="24" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"><path d="M12,1A11,11,0,1,0,23,12,11,11,0,0,0,12,1Zm0,19a8,8,0,1,1,8-8A8,8,0,0,1,12,20Z" opacity=".25"/>
    <path d="M12,4a8,8,0,0,1,7.89,6.7A1.53,1.53,0,0,0,21.38,12h0a1.5,1.5,0,0,0,1.48-1.75,11,11,0,0,0-21.72,0A1.5,1.5,0,0,0,2.62,12h0a1.53,1.53,0,0,0,1.49-1.3A8,8,0,0,1,12,4Z"><animateTransform attributeName="transform" type="rotate" dur="0.75s" values="0 12 12;360 12 12" repeatCount="indefinite"/></path>
  </svg>

  <div v-else-if="name === 'blank'" class="h-6 w-6" />
</template>

<script lang="ts" setup>
import SvgIcon from '@jamescoyle/vue-icon'
import { mdiTimelapse, mdiClockTimeEightOutline, mdiSourceBranch, mdiSourcePull, mdiSourceMerge, mdiGestureTap, mdiTagOutline, mdiPackageVariant, mdiSourceCommit, mdiArrowLeft, mdiGithub, mdiGit, mdiCog, mdiTrashCanOutline, mdiPlay, mdiStop, mdiCloseThick, mdiRadioboxBlank, mdiRadioboxIndeterminateVariant, mdiMinusCircle, mdiCheckCircle, mdiAlert, mdiAlertOutline, mdiAlertCircle, mdiGitlab, mdiBitbucket, mdiHelpCircleOutline, mdiPlus, mdiFormatListBulleted, mdiBandage, mdiPower, mdiChevronRight, mdiCloseCircleOutline, mdiPencil, mdiDownloadCircle, mdiAlarm, mdiDownload, mdiDownloadOff, mdiPause, mdiClose } from '@mdi/js'
import { siForgejo, siGitea } from 'simple-icons';


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
  | 'times';

defineProps<{
  name: IconNames;
}>();
</script>
