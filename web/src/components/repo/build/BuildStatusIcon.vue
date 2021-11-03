<template>
  <div v-if="build" class="flex items-center justify-center">
    <Icon
      :name="`status-${build.status}`"
      :class="{
        'text-yellow-400': build.status === 'pending',
        'text-red-400': buildStatusColors[build.status] === 'red',
        'text-gray-400': buildStatusColors[build.status] === 'gray',
        'text-lime-400': buildStatusColors[build.status] === 'green',
        'text-blue-400': buildStatusColors[build.status] === 'blue',
        [buildStatusAnimations[build.status]]: true,
      }"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import { Build } from '~/lib/api/types';

import { buildStatusAnimations, buildStatusColors } from './build-status';

export default defineComponent({
  name: 'BuildStatusIcon',

  components: {
    Icon,
  },

  props: {
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup() {
    return { buildStatusColors, buildStatusAnimations };
  },
});
</script>
