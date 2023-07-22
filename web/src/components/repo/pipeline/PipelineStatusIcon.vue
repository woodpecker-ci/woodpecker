<template>
  <div
    class="flex items-center justify-center"
    :title="$t('repo.pipeline.status.status', { status: $t(`repo.pipeline.status.${status}`) })"
  >
    <Icon
      :name="`status-${status}`"
      :class="{
        'text-wp-pipeline-error-100': pipelineStatusColors[status] === 'red',
        'text-wp-pipeline-neutral-100': pipelineStatusColors[status] === 'gray',
        'text-wp-pipeline-ok-100': pipelineStatusColors[status] === 'green',
        'text-wp-pipeline-info-100': pipelineStatusColors[status] === 'blue',
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
