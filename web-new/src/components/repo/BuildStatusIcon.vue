<template>
  <div v-if="build" class="flex items-center justify-center">
    <Icon
      :name="`status-${build.status}`"
      :class="{
        'text-yellow-400': build.status === 'pending',
        'text-status-red': buildStatusColors[build.status] === 'red',
        'text-status-gray': buildStatusColors[build.status] === 'gray',
        'text-status-green': buildStatusColors[build.status] === 'green',
        'text-status-blue': buildStatusColors[build.status] === 'blue',
        [buildStatusAnimations[build.status]]: true,
      }"
    />
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';
import { Build } from '~/lib/api/types';
import Icon from '~/components/atomic/Icon.vue';
import { buildStatusColors, buildStatusAnimations } from './build-status';

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
