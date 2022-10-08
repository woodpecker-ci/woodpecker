<template>
  <button
    class="flex rounded-full w-8 h-8 bg-opacity-30 hover:bg-opacity-50 bg-white items-center justify-center cursor-pointer text-white"
    :class="{
      spinner: activeBuilds.length !== 0,
    }"
    type="button"
    @click="toggle"
  >
    <div class="spinner-ring ring1" />
    <div class="spinner-ring ring2" />
    <div class="spinner-ring ring3" />
    <div class="spinner-ring ring4" />
    {{ activeBuilds.length || 0 }}
  </button>
</template>

<script lang="ts">
import { defineComponent, onMounted } from 'vue';

import useBuildFeed from '~/compositions/useBuildFeed';

export default defineComponent({
  name: 'ActiveBuilds',

  setup() {
    const buildFeed = useBuildFeed();

    onMounted(() => {
      buildFeed.load();
    });

    return buildFeed;
  },
});
</script>

<style scoped>
.spinner .spinner-ring {
  animation: spinner 1.2s cubic-bezier(0.5, 0, 0.5, 1) infinite;
  border-color: #fff transparent transparent transparent;
  @apply w-8 h-8 border-2 rounded-full m-4 absolute;
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
