<template>
  <div class="flex bg-green">
    <div class="flex text-white items-center p-4 m-auto w-full">
      <router-link :to="{ name: 'home' }" class="relative">
        <img class="-mt-3 w-8" src="../../../assets/logo.svg" />
        <span class="absolute -bottom-4 text-xs">{{ version }}</span>
      </router-link>
      <a :href="docsUrl" target="_blank" class="ml-8">Docs</a>
      <!-- <router-link v-if="user && user.admin" :to="{ name: 'admin' }" class="ml-8">Administration</router-link> -->
      <router-link :to="{ name: 'repos' }" class="ml-8">Repositories</router-link>
      <router-link
        v-if="$route.matched.some(({ name }) => name === 'repo-wrapper')"
        :to="{ name: 'repo-settings' }"
        class="ml-8"
        >Repo-Settings</router-link
      >
      <div class="flex ml-auto items-center">
        <ActiveBuilds />
        <router-link v-if="user" :to="{ name: 'user' }" class="ml-4">
          <img v-if="user && user.avatar_url" class="w-8" :src="`${user.avatar_url}&s=32`" />
        </router-link>
        <Button v-else text="Login" @click="doLogin" />
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import useConfig from '~/compositions/useConfig';
import Button from '~/components/atomic/Button.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import useAuthentication from '~/compositions/useAuthentication';
import ActiveBuilds from './ActiveBuilds.vue';

export default defineComponent({
  name: 'NavBar',

  components: { Button, FluidContainer, ActiveBuilds },

  setup() {
    const config = useConfig();
    const authentication = useAuthentication();
    const docsUrl = window.WOODPECKER_DOCS;

    function doLogin() {
      authentication.authenticate();
    }

    return { user: authentication.user, doLogin, docsUrl, version: config.version };
  },
});
</script>
