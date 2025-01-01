<template>
  <div class="mx-auto w-full p-4" :class="{ 'max-w-5xl': !fullWidth && !fillWidth, 'md:px-0': fullWidth }">
    <slot />
  </div>
</template>

<script setup lang="ts">
defineProps<{
  fullWidth?: boolean;
  fillWidth?: boolean;
}>();
</script>
