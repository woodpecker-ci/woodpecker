<template>
  <NavbarIcon :title="$t('pipeline_feed')" class="!p-1.5 relative" @click="toggle">
    <div v-if="activePipelines.length > 0" class="spinner">
      <div class="spinner-ring ring1" />
      <div class="spinner-ring ring2" />
      <div class="spinner-ring ring3" />
      <div class="spinner-ring ring4" />
    </div>
    <div
      class="flex items-center justify-center h-full w-full font-bold bg-white bg-opacity-15 dark:bg-black dark:bg-opacity-10 rounded-full"
    >
      {{ activePipelines.length || 0 }}
    </div>
  </NavbarIcon>
</template>

<script lang="ts">
import { defineComponent, onMounted } from 'vue';

import usePipelineFeed from '~/compositions/usePipelineFeed';

import NavbarIcon from './NavbarIcon.vue';

export default defineComponent({
  name: 'ActivePipelines',

  components: { NavbarIcon },

  setup() {
    const pipelineFeed = usePipelineFeed();

    onMounted(() => {
      pipelineFeed.load();
    });

    return pipelineFeed;
  },
});
</script>

<style scoped>
.spinner {
  @apply absolute top-0 bottom-0 left-0 right-0;
}
.spinner .spinner-ring {
  animation: spinner 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;
  border-color: #fff transparent transparent transparent;
  @apply border-3 rounded-full absolute top-1.5 bottom-1.5 left-1.5 right-1.5;
}
.spinner .ring1 {
  animation-delay: -0.45s;
}
.spinner .ring2 {
  animation-delay: -0.3s;
}
.spinner .ring3 {
  animation-delay: -0.15s;
}
@keyframes spinner {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}
</style>
