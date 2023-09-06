<template>
  <div class="w-full p-4 mx-auto" :class="{ 'max-w-5xl': !fullWidth, 'md:px-0': fullWidth }">
    <slot />
  </div>
</template>

<script setup lang="ts">
export interface Props {
  fullWidth?: boolean;
}

defineProps<Props>();
</script>
