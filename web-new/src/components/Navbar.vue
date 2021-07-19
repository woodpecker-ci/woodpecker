<template>
  <div class="flex bg-green">
    <div class="flex w-full max-w-5xl p-4 lg:px-0 m-auto text-white items-center">
      <router-link :to="{ name: 'home' }"><img class="w-8" src="../assets/logo.svg" /></router-link>
      <router-link :to="{ name: 'repos' }" class="ml-8">Projects</router-link>
      <a href="https://woodpecker.laszlo.cloud/" target="_blank" class="ml-8">Docs</a>
      <div class="ml-auto">
        <img v-if="user && user.avatar_url" class="ml-auto w-8" :src="`${user.avatar_url}&s=32`" />
        <Button v-else text="Login" @click="doLogin" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import useConfig from '~/compositions/useConfig';
import Button from '~/components/atomic/Button.vue';
import { authenticate } from '~/compositions/useAuthentication';

export default defineComponent({
  name: 'NavBar',

  components: { Button },

  setup() {
    const config = useConfig();
    const user = config.user;

    function doLogin() {
      authenticate();
    }

    return { user, doLogin };
  },
});
</script>
