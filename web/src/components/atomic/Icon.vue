<template>
  <i-ic-sharp-timelapse v-if="name === 'duration'" class="h-6 w-6" />
  <i-mdi-clock-time-eight-outline v-else-if="name === 'since'" class="h-6 w-6" />
  <i-mdi-source-branch v-else-if="name === 'push'" class="h-6 w-6" />
  <i-mdi-source-pull v-else-if="name === 'pull_request'" class="h-6 w-6" />
  <i-mdi-gesture-tap v-else-if="name === 'manual-pipeline'" class="h-6 w-6" />
  <i-mdi-tag-outline v-else-if="name === 'tag'" class="h-6 w-6" />
  <i-clarity-deploy-line v-else-if="name === 'deployment'" class="h-6 w-6" />
  <i-mdisource-commit v-else-if="name === 'commit'" class="h-6 w-6" />
  <i-iconoir-arrow-left v-else-if="name === 'back'" class="w-8 h-8" />
  <i-mdi-github v-else-if="name === 'github'" class="h-8 w-8" />
  <i-teenyicons-git-solid v-else-if="name === 'repo'" class="h-8 w-8" />
  <i-clarity-settings-solid v-else-if="name === 'settings'" class="w-8 h-8" />
  <i-gg-trash v-else-if="name === 'trash'" class="h-6 w-6" />
  <i-bi-play-circle-fill v-else-if="name === 'status-blocked'" class="h-6 w-6" />
  <i-bi-stop-circle-fill v-else-if="name === 'status-declined'" class="h-6 w-6" />
  <i-bi-x-circle-fill
    v-else-if="name === 'status-failure' || name === 'status-error' || name === 'status-killed'"
    class="h-6 w-6"
  />
  <i-bi-circle v-else-if="name === 'status-pending'" class="h-6 w-6" />
  <i-carbon-in-progress v-else-if="name === 'status-running' || name === 'status-started'" class="h-6 w-6" />
  <i-bi-slash-circle-fill v-else-if="name === 'status-skipped'" class="h-6 w-6" />
  <i-bi-check-circle-fill v-else-if="name === 'status-success'" class="h-6 w-6" />
  <i-bi-x v-else-if="name === 'linter-error'" class="h-6 w-6" />
  <i-simple-icons-gitea v-else-if="name === 'gitea'" class="h-8 w-8" />
  <i-ph-gitlab-logo-simple-fill v-else-if="name === 'gitlab'" class="h-8 w-8" />
  <i-mdi-bitbucket v-else-if="name === 'bitbucket'" class="h-8 w-8" />
  <i-vaadin-question-circle-o v-else-if="name === 'question'" class="h-6 w-6" />
  <i-ic-twotone-add v-else-if="name === 'plus'" class="h-6 w-6" />
  <i-mdi-format-list-bulleted v-else-if="name === 'list'" class="h-6 w-6" />
  <i-mdi-loading v-else-if="name === 'loading'" class="h-6 w-6" />
  <i-ic-baseline-dark-mode v-else-if="name === 'dark'" class="h-6 w-6" />
  <i-ic-round-light-mode v-else-if="name === 'light'" class="h-6 w-6" />
  <i-mdi-sync v-else-if="name === 'sync'" class="h-6 w-6" />
  <i-ic-baseline-healing v-else-if="name === 'heal'" class="h-6 w-6" />
  <i-bx-bx-power-off v-else-if="name === 'turn-off'" class="h-6 w-6" />
  <i-mdi-chevron-right v-else-if="name === 'chevron-right'" class="h-6 w-6" />
  <i-carbon-close-outline v-else-if="name === 'close'" class="h-6 w-6" />
  <i-ic-baseline-edit v-else-if="name === 'edit'" class="h-6 w-6" />
  <i-ic-baseline-download-for-offline v-else-if="name === 'download'" class="h-6 w-6" />
  <i-icon-park-outline-alarm-clock v-else-if="name === 'stopwatch'" class="h-6 w-6" />
  <i-ic-baseline-file-download v-else-if="name === 'auto-scroll'" class="h-6 w-6" />
  <i-ic-baseline-file-download-off v-else-if="name === 'auto-scroll-off'" class="h-6 w-6" />
  <i-teenyicons-refresh-outline v-else-if="name === 'refresh'" class="h-6 w-6" />
  <i-ic-baseline-play-arrow v-else-if="name === 'play'" class="h-6 w-6" />
  <i-ic-baseline-pause v-else-if="name === 'pause'" class="h-6 w-6" />
  <div v-else-if="name === 'blank'" class="h-6 w-6" />
</template>

<script lang="ts" setup>
export type IconNames =
  | 'duration'
  | 'since'
  | 'push'
  | 'pull_request'
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
  | 'linter-error';

defineProps<{
  name: IconNames;
}>();
</script>
