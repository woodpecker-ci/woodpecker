<template>
  <div
    class="flex gap-2 items-center bg-wp-control-error-100 border-l-6 border border-solid border-wp-control-error-300 text-white p-2 rounded-md"
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
