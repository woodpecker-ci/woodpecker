<template>
  <div class="flex bg-green">
    <div class="flex p-4 w-full max-w-5xl m-auto text-white items-center">
      <router-link :to="{ name: 'home' }" class="ml-8"><img class="w-8" src="../assets/logo.svg" /></router-link>
      <router-link :to="{ name: 'projects' }" class="ml-8">Projects</router-link>
      <a href="https://woodpecker.laszlo.cloud/" target="_blank" class="ml-8">Docs</a>
      <img class="ml-auto w-8" src="https://avatars.githubusercontent.com/u/6918444?s=32&v=4" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted } from 'vue';

import useApiClient from '~/compositions/useApiClient';

export default defineComponent({
  name: 'NavBar',

  setup() {
    const apiClient = useApiClient();

    onMounted(async () => {
      const user = await apiClient.getSelf();
      console.log('huhu', user);
    });
  },
});
</script>
