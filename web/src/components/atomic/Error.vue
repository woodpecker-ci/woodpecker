<template>
  <div
    class="flex items-center gap-2 border-wp-state-error-300 bg-wp-state-error-100 p-2 border border-l-6 border-solid rounded-md text-white"
  >
    <Icon v-if="!textOnly" name="error" />
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
