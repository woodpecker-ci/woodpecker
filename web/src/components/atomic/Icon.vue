<!-- cSpell:ignore teenyicons radiobox vaadin twotone iconoir timelapse -->
<template>
  <IIcSharpTimelapse v-if="name === 'duration'" class="h-6 w-6" />
  <IMdiClockTimeEightOutline v-else-if="name === 'since'" class="h-6 w-6" />
  <IMdiSourceBranch v-else-if="name === 'push'" class="h-6 w-6" />
  <IMdiSourcePull v-else-if="name === 'pull-request'" class="h-6 w-6" />
  <IMdiSourceMerge v-else-if="name === 'pull-request-closed'" class="h-6 w-6" />
  <IMdiGestureTap v-else-if="name === 'manual-pipeline'" class="h-6 w-6" />
  <IMdiTagOutline v-else-if="name === 'tag'" class="h-6 w-6" />
  <IClarityDeployLine v-else-if="name === 'deployment'" class="h-6 w-6" />
  <IMdiSourceCommit v-else-if="name === 'commit'" class="h-6 w-6" />
  <IIconoirArrowLeft v-else-if="name === 'back'" class="w-8 h-8" />
  <IMdiGithub v-else-if="name === 'github'" class="h-8 w-8" />
  <ITeenyiconsGitSolid v-else-if="name === 'repo'" class="h-8 w-8" />
  <IClaritySettingsSolid v-else-if="name === 'settings'" class="w-8 h-8" />
  <IGgTrash v-else-if="name === 'trash'" class="h-6 w-6" />
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
  <IBiCheckCircleFill v-else-if="name === 'status-success'" class="h-6 w-6" />
  <IBiExclamationTriangleFill v-else-if="name === 'attention'" class="h-5 w-5" />
  <IBiExclamationTriangle v-else-if="name === 'warning'" class="h-5 w-5" />
  <IMdiErrorOutline v-else-if="name === 'error'" class="h-5 w-5" />
  <ISimpleIconsGitea v-else-if="name === 'gitea'" class="h-8 w-8" />
  <ISimpleIconsForgejo v-else-if="name === 'forgejo'" class="h-8 w-8" />
  <IPhGitlabLogoSimpleFill v-else-if="name === 'gitlab'" class="h-8 w-8" />
  <IMdiBitbucket v-else-if="name === 'bitbucket' || name === 'bitbucket-dc'" class="h-8 w-8" />
  <IVaadinQuestionCircleO v-else-if="name === 'question'" class="h-6 w-6" />
  <IIcTwotoneAdd v-else-if="name === 'plus'" class="h-6 w-6" />
  <IMdiFormatListBulleted v-else-if="name === 'list'" class="h-6 w-6" />
  <IMdiLoading v-else-if="name === 'loading'" class="h-6 w-6" />
  <IIcBaselineDarkMode v-else-if="name === 'dark'" class="h-6 w-6" />
  <IIcRoundLightMode v-else-if="name === 'light'" class="h-6 w-6" />
  <IMdiSync v-else-if="name === 'sync'" class="h-6 w-6" />
  <IIcBaselineHealing v-else-if="name === 'heal'" class="h-6 w-6" />
  <IBxBxPowerOff v-else-if="name === 'turn-off'" class="h-6 w-6" />
  <IMdiChevronRight v-else-if="name === 'chevron-right'" class="h-6 w-6" />
  <ICarbonCloseOutline v-else-if="name === 'close'" class="h-6 w-6" />
  <IIcBaselineEdit v-else-if="name === 'edit'" class="h-6 w-6" />
  <IIcBaselineDownloadForOffline v-else-if="name === 'download'" class="h-6 w-6" />
  <IIconParkOutlineAlarmClock v-else-if="name === 'stopwatch'" class="h-6 w-6" />
  <IIcBaselineFileDownload v-else-if="name === 'auto-scroll'" class="h-6 w-6" />
  <IIcBaselineFileDownloadOff v-else-if="name === 'auto-scroll-off'" class="h-6 w-6" />
  <ITeenyiconsRefreshOutline v-else-if="name === 'refresh'" class="h-6 w-6" />
  <IIcBaselinePlayArrow v-else-if="name === 'play'" class="h-6 w-6" />
  <IIcBaselinePause v-else-if="name === 'pause'" class="h-6 w-6" />
  <ISvgSpinners180RingWithBg v-else-if="name === 'spinner'" class="h-6 w-6" />
  <ILaTimes v-else-if="name === 'times'" class="h-6 w-6" />
  <div v-else-if="name === 'blank'" class="h-6 w-6" />
</template>

<script lang="ts" setup>
import IBiCheckCircleFill from '~icons/bi/check-circle-fill';
import IBiExclamationTriangle from '~icons/bi/exclamation-triangle';
import IBiExclamationTriangleFill from '~icons/bi/exclamation-triangle-fill';
import IBiSlashCircleFill from '~icons/bi/slash-circle-fill';
import IBxBxPowerOff from '~icons/bx/bx-power-off';
import ICarbonCloseOutline from '~icons/carbon/close-outline';
import IClarityDeployLine from '~icons/clarity/deploy-line';
import IClaritySettingsSolid from '~icons/clarity/settings-solid';
import IGgTrash from '~icons/gg/trash';
import IIcBaselineDarkMode from '~icons/ic/baseline-dark-mode';
import IIcBaselineDownloadForOffline from '~icons/ic/baseline-download-for-offline';
import IIcBaselineEdit from '~icons/ic/baseline-edit';
import IIcBaselineFileDownload from '~icons/ic/baseline-file-download';
import IIcBaselineFileDownloadOff from '~icons/ic/baseline-file-download-off';
import IIcBaselineHealing from '~icons/ic/baseline-healing';
import IIcBaselinePause from '~icons/ic/baseline-pause';
import IIcBaselinePlayArrow from '~icons/ic/baseline-play-arrow';
import IIcRoundLightMode from '~icons/ic/round-light-mode';
import IIcSharpTimelapse from '~icons/ic/sharp-timelapse';
import IIcTwotoneAdd from '~icons/ic/twotone-add';
import IIconParkOutlineAlarmClock from '~icons/icon-park-outline/alarm-clock';
import IIconoirArrowLeft from '~icons/iconoir/arrow-left';
import ILaTimes from '~icons/la/times';
import IMdiBitbucket from '~icons/mdi/bitbucket';
import IMdiChevronRight from '~icons/mdi/chevron-right';
import IMdiClockTimeEightOutline from '~icons/mdi/clock-time-eight-outline';
import IMdiCloseThick from '~icons/mdi/close-thick';
import IMdiErrorOutline from '~icons/mdi/error-outline';
import IMdiFormatListBulleted from '~icons/mdi/format-list-bulleted';
import IMdiGestureTap from '~icons/mdi/gesture-tap';
import IMdiGithub from '~icons/mdi/github';
import IMdiLoading from '~icons/mdi/loading';
import IMdiPlay from '~icons/mdi/play';
import IMdiRadioboxBlank from '~icons/mdi/radiobox-blank';
import IMdiRadioboxIndeterminateVariant from '~icons/mdi/radiobox-indeterminate-variant';
import IMdiSourceBranch from '~icons/mdi/source-branch';
import IMdiSourceCommit from '~icons/mdi/source-commit';
import IMdiSourceMerge from '~icons/mdi/source-merge';
import IMdiSourcePull from '~icons/mdi/source-pull';
import IMdiStop from '~icons/mdi/stop';
import IMdiSync from '~icons/mdi/sync';
import IMdiTagOutline from '~icons/mdi/tag-outline';
import IPhGitlabLogoSimpleFill from '~icons/ph/gitlab-logo-simple-fill';
import ISimpleIconsForgejo from '~icons/simple-icons/forgejo';
import ISimpleIconsGitea from '~icons/simple-icons/gitea';
import ISvgSpinners180RingWithBg from '~icons/svg-spinners/180-ring-with-bg';
import ITeenyiconsGitSolid from '~icons/teenyicons/git-solid';
import ITeenyiconsRefreshOutline from '~icons/teenyicons/refresh-outline';
import IVaadinQuestionCircleO from '~icons/vaadin/question-circle-o';

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
  | 'dark'
  | 'light'
  | 'sync'
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
