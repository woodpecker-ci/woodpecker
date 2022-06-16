<template>
  <div class="flex shadow-lg bg-lime-600 text-neutral-content px-2 md:px-8 py-2 dark:bg-dark-gray-900">
    <div class="flex text-white dark:text-gray-500 items-center">
      <router-link :to="{ name: 'home' }" class="relative">
        <img class="-mt-3 w-8" src="../../../assets/logo.svg?url" />
        <span class="absolute -bottom-4 text-xs">{{ version }}</span>
      </router-link>
      <router-link
        v-if="user"
        :to="{ name: 'repos' }"
        class="mx-4 hover:bg-lime-700 dark:hover:bg-gray-600 px-4 py-1 rounded-md"
      >
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="hidden md:flex">{{ $t('repositories') }}</span>
      </router-link>
    </div>
    <div class="flex ml-auto items-center space-x-4 text-white dark:text-gray-500">
      <a
        :href="docsUrl"
        target="_blank"
        class="hover:bg-lime-700 dark:hover:bg-gray-600 px-4 py-1 rounded-md hidden md:flex"
        >Docs</a
      >
      <IconButton
        :icon="darkMode ? 'dark' : 'light'"
        class="!text-white !dark:text-gray-500"
        @click="darkMode = !darkMode"
      />
      <router-link v-if="user" :to="{ name: 'user' }">
        <img v-if="user && user.avatar_url" class="w-8" :src="`${user.avatar_url}`" />
      </router-link>
      <Button v-else :text="$t('login')" @click="doLogin" />
      <ActiveBuilds v-if="user" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { useRoute } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import { useDarkMode } from '~/compositions/useDarkMode';

import ActiveBuilds from './ActiveBuilds.vue';

export default defineComponent({
  name: 'Navbar',

  components: { Button, ActiveBuilds, IconButton },

  setup() {
    const config = useConfig();
    const route = useRoute();
    const authentication = useAuthentication();
    const { darkMode } = useDarkMode();
    const docsUrl = window.WOODPECKER_DOCS;

    function doLogin() {
      authentication.authenticate(route.fullPath);
    }

    const version = config.version?.startsWith('next') ? 'next' : config.version;

    return { darkMode, user: authentication.user, doLogin, docsUrl, version };
  },
});
</script>
