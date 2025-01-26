<template>
  <IconButton
    :title="pipelineCount > 0 ? `${$t('pipeline_feed')} (${pipelineCount})` : $t('pipeline_feed')"
    class="active-pipelines-toggle relative p-1.5! text-current"
    @click="toggle"
  >
    <div v-if="pipelineCount > 0" class="spinner" />
    <div
      class="z-0 flex h-full w-full items-center justify-center rounded-md bg-white bg-opacity-15 font-bold dark:bg-black dark:bg-opacity-10"
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
@reference '~/tailwind.css';

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
