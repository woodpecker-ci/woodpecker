<template>
  <a :href="`${docsBaseUrl}${url}`" target="_blank" class="text-blue-500 hover:text-blue-600 cursor-pointer mt-1"
    ><Icon name="question" class="!w-4 !h-4"
  /></a>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import Icon from '~/components/atomic/Icon.vue';

export default defineComponent({
  name: 'DocsLink',

  components: {
    Icon,
  },

  props: {
    url: {
      type: String,
      required: true,
    },
  },

  setup() {
    const docsBaseUrl = window.WOODPECKER_DOCS;

    return { docsBaseUrl };
  },
});
</script>
