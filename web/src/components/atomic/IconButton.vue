<template>
  <router-link v-if="to" :to="to" :title="title" :aria-label="title" class="w-8 icon-button">
    <slot>
      <Icon v-if="icon" :name="icon" />
    </slot>
  </router-link>
  <a
    v-else-if="href"
    :href="href"
    :title="title"
    :aria-label="title"
    class="icon-button"
    target="_blank"
    rel="noopener noreferrer"
  >
    <slot>
      <Icon v-if="icon" :name="icon" />
    </slot>
  </a>
  <button v-else :disabled="disabled" class="icon-button" type="button" :title="title" :aria-label="title">
    <slot>
      <Icon v-if="icon" :name="icon" />
    </slot>
    <div v-if="isLoading" class="top-0 right-0 bottom-0 left-0 absolute flex justify-center items-center">
      <Icon name="loading" class="animate-spin" />
    </div>
  </button>
</template>

<script lang="ts" setup>
import type { RouteLocationRaw } from 'vue-router';

import Icon, { type IconNames } from '~/components/atomic/Icon.vue';

defineProps<{
  icon?: IconNames;
  disabled?: boolean;
  to?: RouteLocationRaw;
  isLoading?: boolean;
  title?: string;
  href?: string;
}>();
</script>

<style scoped>
.icon-button {
  @apply relative flex justify-center items-center bg-transparent disabled:opacity-50 px-1 py-1 rounded-md disabled:cursor-not-allowed overflow-hidden hover-effect;
}
</style>
