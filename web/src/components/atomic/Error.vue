<template>
  <div
    class="border-wp-error-200 bg-wp-error-100 flex items-center gap-2 rounded-md border border-l-4 border-solid p-2 text-white"
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
