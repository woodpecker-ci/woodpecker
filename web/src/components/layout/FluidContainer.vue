<template>
  <div class="w-full p-2 md:p-4 lg:px-0 mx-auto" :class="{ 'max-w-5xl': !fullWidth }">
    <slot />
  </div>
</template>

<script setup lang="ts">
export interface Props {
  fullWidth?: boolean;
}

defineProps<Props>();
</script>
