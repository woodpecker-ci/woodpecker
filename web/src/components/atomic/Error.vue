<template>
  <div
    class="flex items-center gap-2 rounded-md border border-l-6 border-solid border-wp-error-300 bg-wp-error-100 p-2 text-white dark:border-wp-error-300 dark:bg-wp-error-200"
  >
    <Icon v-if="!textOnly" name="alert" />
    <slot>
      <span class="whitespace-pre">{{ text }}</span>
    </slot>
  </div>
</template>

<script lang="ts" setup>
import Icon from './Icon.vue';

defineProps<{
  textOnly?: boolean;
  text?: string;
}>();
</script>
