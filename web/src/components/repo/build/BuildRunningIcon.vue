<template>
  <WoodpeckerIcon class="woodpecker h-16" />
</template>

<script lang="ts">
import { defineComponent } from 'vue';

// eslint-disable-next-line import/no-relative-parent-imports
import WoodpeckerIcon from '../../../assets/woodpecker.svg?component';

export default defineComponent({
  name: 'BuildRunningIcon',

  components: {
    WoodpeckerIcon,
  },
});
</script>

<style scoped>
@keyframes peck {
  0% {
    transform: rotate(5deg) translateX(5%);
  }
  10% {
    transform: rotate(-5deg) translateX(-15%);
  }
  20% {
    transform: rotate(5deg) translateX(5%);
  }
  30% {
    transform: rotate(-5deg) translateX(-15%);
  }
  100% {
    transform: rotate(5deg) translateX(5%);
  }
}

.woodpecker ::v-deep(path) {
  animation: peck 1s ease infinite;
  @apply fill-gray-600 dark:fill-gray-500;
}
</style>
