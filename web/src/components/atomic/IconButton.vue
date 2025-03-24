<template>
  <router-link v-if="to" :to="to" :title="title" :aria-label="title" class="h-8 w-8" :class="btnClass">
    <slot>
      <Icon v-if="icon" :name="icon" />
    </slot>
  </router-link>
  <a
    v-else-if="href"
    :href="href"
    :title="title"
    :aria-label="title"
    :class="btnClass"
    target="_blank"
    rel="noopener noreferrer"
  >
    <slot>
      <Icon v-if="icon" :name="icon" />
    </slot>
  </a>
  <button v-else :disabled="disabled" :class="btnClass" type="button" :title="title" :aria-label="title">
    <slot>
      <Icon v-if="icon" :name="icon" />
    </slot>
    <div v-if="isLoading" class="absolute top-0 right-0 bottom-0 left-0 flex items-center justify-center">
      <Icon name="loading" class="animate-spin" />
    </div>
  </button>
</template>

<script lang="ts" setup>
import type { RouteLocationRaw } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';
import type { IconNames } from '~/components/atomic/Icon.vue';

defineProps<{
  icon?: IconNames;
  disabled?: boolean;
  to?: RouteLocationRaw;
  isLoading?: boolean;
  title?: string;
  href?: string;
}>();

const btnClass =
  'hover-effect relative flex cursor-pointer items-center justify-center overflow-hidden rounded-md bg-transparent px-1 py-1 disabled:cursor-not-allowed disabled:opacity-50';
</script>
