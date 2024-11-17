<template>
  <div
    class="flex gap-4 items-center text-gray-700 font-bold rounded-md p-4 border border-solid border-l-6 border-wp-hint-warn-200 bg-wp-hint-warn-100"
  >
    <Icon v-if="!textOnly" name="warning" class="flex-shrink-0" />
    <slot>
      <span class="whitespace-pre-wrap">{{ text }}</span>
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
