<template>
  <div
    class="flex items-center justify-center"
    :title="$t('repo.pipeline.status.status', { status: $t(`repo.pipeline.status.${status}`) })"
  >
    <Icon
      :name="`status-${status}`"
      :class="{
        'text-red-400': pipelineStatusColors[status] === 'red',
        'text-gray-400': pipelineStatusColors[status] === 'gray',
        'text-woodpecker-400': pipelineStatusColors[status] === 'green',
        'text-blue-400': pipelineStatusColors[status] === 'blue',
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
