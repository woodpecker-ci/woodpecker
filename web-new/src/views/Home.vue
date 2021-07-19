<template>
  <div class="max-w-5xl w-full h-full m-auto">Hello</div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import useConfig from '~/compositions/useConfig';

export default defineComponent({
  name: 'Home',

  setup() {
    const user = useConfig().user;

    return { user };
  },
});
</script>
