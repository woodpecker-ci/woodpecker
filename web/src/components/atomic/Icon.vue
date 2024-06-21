<!-- cSpell:ignore radiobox timelapse -->
<template>
  <IMdiTimelapse v-if="name === 'duration'" class="h-6 w-6" />
  <IMdiClockTimeEightOutline v-else-if="name === 'since'" class="h-6 w-6" />
  <IMdiSourceBranch v-else-if="name === 'push'" class="h-6 w-6" />
  <IMdiSourcePull v-else-if="name === 'pull-request'" class="h-6 w-6" />
  <IMdiSourceMerge v-else-if="name === 'pull-request-closed'" class="h-6 w-6" />
  <IMdiGestureTap v-else-if="name === 'manual-pipeline'" class="h-6 w-6" />
  <IMdiTagOutline v-else-if="name === 'tag'" class="h-6 w-6" />
  <IMdiPackageVariant v-else-if="name === 'deployment'" class="h-6 w-6" />
  <IMdiSourceCommit v-else-if="name === 'commit'" class="h-6 w-6" />
  <IMdiArrowLeft v-else-if="name === 'back'" class="w-8 h-8" />
  <IMdiGithub v-else-if="name === 'github'" class="h-8 w-8" />
  <IMdiGit v-else-if="name === 'repo'" class="h-8 w-8" />
  <IMdiSettings v-else-if="name === 'settings'" class="w-8 h-8" />
  <IMdiTrashCanOutline v-else-if="name === 'trash'" class="h-6 w-6" />
  <IMdiPlay v-else-if="name === 'status-blocked'" class="h-6 w-6" />
  <IMdiStop v-else-if="name === 'status-declined'" class="h-6 w-6" />
  <IMdiCloseThick
    v-else-if="name === 'status-failure' || name === 'status-error' || name === 'status-killed'"
    class="h-6 w-6"
  />
  <IMdiRadioboxBlank v-else-if="name === 'status-pending'" class="h-6 w-6" />
  <IMdiRadioboxIndeterminateVariant
    v-else-if="name === 'status-running' || name === 'status-started'"
    class="h-6 w-6"
  />
  <IBiSlashCircleFill v-else-if="name === 'status-skipped'" class="h-6 w-6" />
  <IMdiCheckCircle v-else-if="name === 'status-success'" class="h-6 w-6" />
  <IClarityExclamationTriangleSolid v-else-if="name === 'attention'" class="h-6 w-6" />
  <IClarityExclamationTriangleLine v-else-if="name === 'warning'" class="h-6 w-6" />
  <IMdiErrorOutline v-else-if="name === 'error'" class="h-5 w-5" />
  <ISimpleIconsGitea v-else-if="name === 'gitea'" class="h-8 w-8" />
  <ISimpleIconsForgejo v-else-if="name === 'forgejo'" class="h-8 w-8" />
  <IMdiGitlab v-else-if="name === 'gitlab'" class="h-8 w-8" />
  <IMdiBitbucket v-else-if="name === 'bitbucket' || name === 'bitbucket-dc'" class="h-8 w-8" />
  <IMdiQuestionMarkCircleOutline v-else-if="name === 'question'" class="h-6 w-6" />
  <IMdiAdd v-else-if="name === 'plus'" class="h-6 w-6" />
  <IMdiFormatListBulleted v-else-if="name === 'list'" class="h-6 w-6" />
  <IMdiPlaster v-else-if="name === 'heal'" class="h-6 w-6" />
  <IMdiPower v-else-if="name === 'turn-off'" class="h-6 w-6" />
  <IMdiChevronRight v-else-if="name === 'chevron-right'" class="h-6 w-6" />
  <IMdiCloseCircleOutline v-else-if="name === 'close'" class="h-6 w-6" />
  <IMdiEdit v-else-if="name === 'edit'" class="h-6 w-6" />
  <IMdiDownloadCircle v-else-if="name === 'download'" class="h-6 w-6" />
  <IMdiAlarmClock v-else-if="name === 'stopwatch'" class="h-6 w-6" />
  <IMdiDownload v-else-if="name === 'auto-scroll'" class="h-6 w-6" />
  <IMdiDownloadOff v-else-if="name === 'auto-scroll-off'" class="h-6 w-6" />
  <IMdiPlayArrow v-else-if="name === 'play'" class="h-6 w-6" />
  <IMdiPause v-else-if="name === 'pause'" class="h-6 w-6" />
  <ISvgSpinners180RingWithBg v-else-if="name === 'spinner'" class="h-6 w-6" />
  <IMdiClose v-else-if="name === 'times'" class="h-6 w-6" />
  <div v-else-if="name === 'blank'" class="h-6 w-6" />
</template>

<script lang="ts" setup>
import IBiSlashCircleFill from '~icons/bi/slash-circle-fill';
import IMdiPackageVariant from '~icons/mdi/package-variant';
import IClarityExclamationTriangleLine from '~icons/clarity/exclamation-triangle-line';
import IClarityExclamationTriangleSolid from '~icons/clarity/exclamation-triangle-solid';
import IMdiAdd from '~icons/mdi/add';
import IMdiAlarmClock from '~icons/mdi/alarm-clock';
import IMdiArrowLeft from '~icons/mdi/arrow-left';
import IMdiBitbucket from '~icons/mdi/bitbucket';
import IMdiCheckCircle from '~icons/mdi/check-circle';
import IMdiChevronRight from '~icons/mdi/chevron-right';
import IMdiClockTimeEightOutline from '~icons/mdi/clock-time-eight-outline';
import IMdiClose from '~icons/mdi/close';
import IMdiCloseCircleOutline from '~icons/mdi/close-circle-outline';
import IMdiCloseThick from '~icons/mdi/close-thick';
import IMdiDownload from '~icons/mdi/download';
import IMdiDownloadCircle from '~icons/mdi/download-circle';
import IMdiDownloadOff from '~icons/mdi/download-off';
import IMdiEdit from '~icons/mdi/edit';
import IMdiErrorOutline from '~icons/mdi/error-outline';
import IMdiFormatListBulleted from '~icons/mdi/format-list-bulleted';
import IMdiGestureTap from '~icons/mdi/gesture-tap';
import IMdiGit from '~icons/mdi/git';
import IMdiGithub from '~icons/mdi/github';
import IMdiGitlab from '~icons/mdi/gitlab';
import IMdiPause from '~icons/mdi/pause';
import IMdiPlaster from '~icons/mdi/plaster';
import IMdiPlay from '~icons/mdi/play';
import IMdiPlayArrow from '~icons/mdi/play-arrow';
import IMdiPower from '~icons/mdi/power';
import IMdiQuestionMarkCircleOutline from '~icons/mdi/question-mark-circle-outline';
import IMdiRadioboxBlank from '~icons/mdi/radiobox-blank';
import IMdiRadioboxIndeterminateVariant from '~icons/mdi/radiobox-indeterminate-variant';
import IMdiSettings from '~icons/mdi/settings';
import IMdiSourceBranch from '~icons/mdi/source-branch';
import IMdiSourceCommit from '~icons/mdi/source-commit';
import IMdiSourceMerge from '~icons/mdi/source-merge';
import IMdiSourcePull from '~icons/mdi/source-pull';
import IMdiStop from '~icons/mdi/stop';
import IMdiTagOutline from '~icons/mdi/tag-outline';
import IMdiTimelapse from '~icons/mdi/timelapse';
import IMdiTrashCanOutline from '~icons/mdi/trash-can-outline';
import ISimpleIconsForgejo from '~icons/simple-icons/forgejo';
import ISimpleIconsGitea from '~icons/simple-icons/gitea';
import ISvgSpinners180RingWithBg from '~icons/svg-spinners/180-ring-with-bg';

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
