<template>
  <div class="flex flex-col mt-2 mb-4">
    <div class="flex items-center text-color font-bold mb-2">
      <label v-if="label" v-bind="$attrs">{{ label }}</label>
      <DocsLink v-if="docsUrl" :url="docsUrl" class="ml-2" />
    </div>
    <slot />
    <div v-if="$slots['description']" class="ml-1 text-color-alt">
      <slot name="description" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import DocsLink from '~/components/atomic/DocsLink.vue';

export default defineComponent({
  name: 'InputField',

  components: { DocsLink },

  inheritAttrs: false,

  props: {
    label: {
      type: String,
      default: null,
    },

    docsUrl: {
      type: String,
      default: null,
    },
  },
});
</script>
