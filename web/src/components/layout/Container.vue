<template>
  <div class="mx-auto p-4 w-full" :class="{ 'max-w-6xl': !fullWidth && !fillWidth, 'md:px-0': fullWidth }">
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  fullWidth?: boolean;
  fillWidth?: boolean;
}>();
</script>
