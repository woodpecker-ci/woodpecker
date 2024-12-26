<template>
  <div
    class="flex items-center justify-center"
    :title="$t('repo.pipeline.status.status', { status: statusDescriptions[status] })"
  >
    <Icon
      :name="service ? 'settings' : `status-${status}`"
      size="1.5rem"
      :class="{
        'text-wp-state-error-100': pipelineStatusColors[status] === 'red',
        'text-wp-state-neutral-100': pipelineStatusColors[status] === 'gray',
        'text-wp-state-ok-100': pipelineStatusColors[status] === 'green',
        'text-wp-state-info-100': pipelineStatusColors[status] === 'blue',
        'text-wp-state-warn-100': pipelineStatusColors[status] === 'orange',
        'animate-spin': service && pipelineStatusColors[status] === 'blue',
      }"
    />
  </div>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import type { PipelineStatus } from '~/lib/api/types';

import { pipelineStatusColors } from './pipeline-status';

defineProps<{
  status: PipelineStatus;
  service?: boolean;
}>();

const { t } = useI18n();

const statusDescriptions = {
  blocked: t('repo.pipeline.status.blocked'),
  declined: t('repo.pipeline.status.declined'),
  error: t('repo.pipeline.status.error'),
  failure: t('repo.pipeline.status.failure'),
  killed: t('repo.pipeline.status.killed'),
  pending: t('repo.pipeline.status.pending'),
  running: t('repo.pipeline.status.running'),
  skipped: t('repo.pipeline.status.skipped'),
  started: t('repo.pipeline.status.started'),
  success: t('repo.pipeline.status.success'),
} satisfies {
  // eslint-disable-next-line no-unused-vars
  [_ in PipelineStatus]: string;
};
</script>
