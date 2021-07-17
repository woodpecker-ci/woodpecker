<template>
  <div class="not-found">
    <a :href="loginUrl" target="_self">Login</a>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

export default defineComponent({
  name: 'Login',

  props: {
    origin: {
      type: String as PropType<string | null>,
      default: null,
    },
  },

  setup(props) {
    const loginUrl = `/login?url=${props.origin || ''}`;

    return {
      loginUrl,
    };
  },
});
</script>
