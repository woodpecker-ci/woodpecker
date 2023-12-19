<template>
  <div
    class="flex items-center justify-center"
    :title="$t('repo.pipeline.status.status', { status: $t(`repo.pipeline.status.${status}`) })"
  >
    <Icon
      :name="`status-${status}`"
      :class="{
        'text-wp-state-error-100': pipelineStatusColors[status] === 'red',
        'text-wp-state-neutral-100': pipelineStatusColors[status] === 'gray',
        'text-wp-state-ok-100': pipelineStatusColors[status] === 'green',
        'text-wp-state-info-100': pipelineStatusColors[status] === 'blue',
        'text-wp-state-warn-100': pipelineStatusColors[status] === 'orange',
      }"
    />
  </div>
</template>

<script lang="ts" setup>
import Icon from '~/components/atomic/Icon.vue';
import { PipelineStatus } from '~/lib/api/types';

import { pipelineStatusColors } from './pipeline-status';

defineProps<{
  status: PipelineStatus;
}>();
</script>
