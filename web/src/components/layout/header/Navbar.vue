<template>
  <div class="flex shadow-lg bg-lime-600 text-neutral-content px-8 py-3">
    <div class="flex text-white items-center">
      <router-link :to="{ name: 'home' }" class="relative">
        <img class="-mt-3 w-8" src="../../../assets/logo.svg" />
        <span class="absolute -bottom-4 text-xs">{{ version }}</span>
      </router-link>
      <router-link v-if="user" :to="{ name: 'repos' }" class="mx-4 hover:bg-lime-700 px-2 rounded-md"
        >Repositories</router-link
      >
    </div>
    <div class="flex ml-auto items-center space-x-4 text-white">
      <a :href="docsUrl" target="_blank" class="mx-4 hover:bg-lime-700 px-2 rounded-md">Docs</a>
      <router-link v-if="user" :to="{ name: 'user' }">
        <img v-if="user && user.avatar_url" class="w-8" :src="`${user.avatar_url}`" />
      </router-link>
      <Button v-else text="Login" @click="doLogin" />
      <ActiveBuilds v-if="user" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';

import Button from '~/components/atomic/Button.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';

import ActiveBuilds from './ActiveBuilds.vue';

export default defineComponent({
  name: 'Navbar',

  components: { Button, ActiveBuilds },

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
