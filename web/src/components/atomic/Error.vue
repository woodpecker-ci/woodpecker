<template>
  <div
    class="border-wp-error-300 dark:border-wp-error-300 bg-wp-error-100 dark:bg-wp-error-200 border-l-6 flex items-center gap-2 rounded-md border border-solid p-2 text-white"
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
