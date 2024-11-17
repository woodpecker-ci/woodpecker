<template>
  <IconButton
    :title="pipelineCount > 0 ? `${$t('pipeline_feed')} (${pipelineCount})` : $t('pipeline_feed')"
    class="!p-1.5 relative text-current active-pipelines-toggle"
    @click="toggle"
  >
    <div v-if="pipelineCount > 0" class="spinner" />
    <div
      class="z-0 flex items-center justify-center h-full w-full font-bold bg-white bg-opacity-15 dark:bg-black dark:bg-opacity-10 rounded-md"
    >
      <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
      {{ pipelineCount > 9 ? '9+' : pipelineCount }}
    </div>
  </IconButton>
</template>

<script lang="ts" setup>
import { computed, onMounted, toRef } from 'vue';

import IconButton from '~/components/atomic/IconButton.vue';
import usePipelineFeed from '~/compositions/usePipelineFeed';

const pipelineFeed = usePipelineFeed();
const activePipelines = toRef(pipelineFeed, 'activePipelines');
const { toggle } = pipelineFeed;
const pipelineCount = computed(() => activePipelines.value.length || 0);

onMounted(async () => {
  await pipelineFeed.load();
});
</script>

<style scoped>
@keyframes rotate {
  100% {
    transform: rotate(1turn);
  }
}
.spinner {
  @apply absolute z-0 inset-1.5 rounded-md;
  overflow: hidden;
}
.spinner::before {
  @apply absolute -z-2 -inset-1 bg-wp-primary-200 dark:bg-wp-primary-300;
  content: '';
  background: #fff;
  clip-path: polygon(100% 0, 50% 50%, 100% 100%);
  animation: rotate 1.5s linear infinite;
}
.spinner::after {
  @apply absolute inset-0.5 bg-blend-darken bg-wp-primary-200 dark:bg-wp-primary-300;
  /*
  The nested border radius needs to be calculated correctly to look right:
  https://www.30secondsofcode.org/css/s/nested-border-radius/
  */
  border-radius: calc(0.375rem - 0.125rem);
  content: '';
}
</style>
