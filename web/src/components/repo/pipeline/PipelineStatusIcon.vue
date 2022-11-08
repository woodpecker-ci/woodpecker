<template>
  <div class="flex items-center justify-center">
    <Icon
      :name="`status-${status}`"
      :class="{
        'text-yellow-400': status === 'pending',
        'text-red-400': pipelineStatusColors[status] === 'red',
        'text-gray-400': pipelineStatusColors[status] === 'gray',
        'text-lime-400': pipelineStatusColors[status] === 'green',
        'text-blue-400': pipelineStatusColors[status] === 'blue',
        [pipelineStatusAnimations[status]]: true,
      }"
    />
  </div>
</template>

<script lang="ts" setup>
import Icon from '~/components/atomic/Icon.vue';
import { PipelineStatus } from '~/lib/api/types';

import { pipelineStatusAnimations, pipelineStatusColors } from './pipeline-status';

defineProps<{
  status: PipelineStatus;
}>();
</script>
