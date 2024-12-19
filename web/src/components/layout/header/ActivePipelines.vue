<template>
  <IconButton
    :title="pipelineCount > 0 ? `${$t('pipeline_feed')} (${pipelineCount})` : $t('pipeline_feed')"
    class="relative text-current active-pipelines-toggle"
    @click="toggle"
  >
    <div v-if="pipelineCount > 0" class="spinner" />
    <div
      class="z-0 flex justify-center items-center bg-white dark:bg-black bg-opacity-15 dark:bg-opacity-10 rounded-md w-full h-full font-bold"
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
  @apply absolute inset-1.5 rounded-md;
  overflow: hidden;
}
.spinner::before {
  @apply absolute bg-wp-primary-200 dark:bg-wp-primary-300;
  content: '';
  left: -50%;
  top: -50%;
  width: 200%;
  height: 200%;
  background-repeat: no-repeat;
  background-size:
    50% 50%,
    50% 50%;
  background-image: linear-gradient(#fff, #fff);
  animation: rotate 1.5s linear infinite;
}
.spinner::after {
  @apply absolute inset-0.5 bg-wp-primary-200 dark:bg-wp-primary-300;
  /*
  The nested border radius needs to be calculated correctly to look right:
  https://www.30secondsofcode.org/css/s/nested-border-radius/
  */
  border-radius: calc(0.375rem - 0.125rem);
  content: '';
}
</style>
