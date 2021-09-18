<template>
  <div class="flex flex-col mt-2 mb-4">
    <label v-if="label" v-bind="$attrs" class="text-gray-400 font-bold mb-2">{{ label }}</label>
    <slot />
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

export default defineComponent({
  name: 'InputField',

  inheritAttrs: false,

  props: {
    label: {
      type: String,
      default: null,
    },
  },
});
</script>
