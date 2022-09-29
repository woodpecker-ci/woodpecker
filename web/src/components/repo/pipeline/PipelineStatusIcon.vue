<template>
  <div v-if="pipeline" class="flex items-center justify-center">
    <Icon
      :name="`status-${pipeline.status}`"
      :class="{
        'text-yellow-400': pipeline.status === 'pending',
        'text-red-400': pipelineStatusColors[pipeline.status] === 'red',
        'text-gray-400': pipelineStatusColors[pipeline.status] === 'gray',
        'text-lime-400': pipelineStatusColors[pipeline.status] === 'green',
        'text-blue-400': pipelineStatusColors[pipeline.status] === 'blue',
        [pipelineStatusAnimations[pipeline.status]]: true,
      }"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import { Pipeline } from '~/lib/api/types';

import { pipelineStatusAnimations, pipelineStatusColors } from './pipeline-status';

export default defineComponent({
  name: 'PipelineStatusIcon',

  components: {
    Icon,
  },

  props: {
    pipeline: {
      type: Object as PropType<Pipeline>,
      required: true,
    },
  },

  setup() {
    return { pipelineStatusColors, pipelineStatusAnimations };
  },
});
</script>
