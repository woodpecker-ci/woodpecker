<template>
  <div
    class="border-wp-hint-warn-200 bg-wp-hint-warn-100 flex items-center gap-4 rounded-md border border-l-4 border-solid p-4 font-bold text-gray-700"
  >
    <Icon v-if="!textOnly" name="alert" class="shrink-0" />
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
