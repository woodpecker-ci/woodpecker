<template>
  <IconButton
    :title="pipelineCount > 0 ? `${$t('pipeline_feed')} (${pipelineCount})` : $t('pipeline_feed')"
    class="active-pipelines-toggle relative p-1.5! text-current"
    @click="toggle"
  >

    <div v-if="true" class="overflow-hidden">
      <span class="absolute bg-wp-primary-200 dark:bg-wp-primary-300 w-2/1 h-2/1 -left-1/2 -top-1/2 bg-[linear-gradient(#fff,_#fff)] bg-no-repeat animate-[spin_1.5s_linear_infinite]"/>
      <span class="absolute inset-0.5 bg-wp-primary-200 dark:bg-wp-primary-300 rounded-[calc(0.375rem_-_0.125rem)]"></span>
    </div>

    <div
      class="z-0 flex h-full w-full items-center justify-center rounded-md bg-white/15 font-bold dark:bg-black/10"
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
