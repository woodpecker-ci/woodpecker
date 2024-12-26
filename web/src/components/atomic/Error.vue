<template>
  <div
    class="bg-wp-control-error-100 border-l-6 border-wp-control-error-300 flex items-center gap-2 rounded-md border border-solid p-2 text-white"
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
