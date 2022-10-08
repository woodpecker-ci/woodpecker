<template>
  <!-- Navbar -->
  <div class="flex shadow-lg dark:shadow-sm bg-lime-600 text-neutral-content p-4 dark:bg-dark-gray-900">
    <!-- Left Links Box -->
    <div class="flex text-white dark:text-gray-400 items-center space-x-2">
      <!-- Logo -->
      <router-link :to="{ name: 'home' }" class="flex flex-col -my-2 px-2">
        <img class="w-8 h-8" src="../../../assets/logo.svg?url" />
        <span class="text-xs">{{ version }}</span>
      </router-link>
      <!-- Repo Link -->
      <router-link v-if="user" :to="{ name: 'repos' }" class="navbar-link navbar-clickable">
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="hidden md:flex">{{ $t('repositories') }}</span>
      </router-link>
      <!-- Docs Link -->
      <a :href="docsUrl" target="_blank" class="navbar-link navbar-clickable hidden md:flex">{{ $t('docs') }}</a>
    </div>
    <!-- Right Icons Box -->
    <div class="flex ml-auto -m-1.5 items-center space-x-2 text-white dark:text-gray-400">
      <!-- Dark Mode Toggle -->
      <NavbarIcon
        :title="$t(darkMode ? 'color_scheme_dark' : 'color_scheme_light')"
        class="navbar-icon navbar-clickable"
        @click="darkMode = !darkMode"
      >
        <i-ic-baseline-dark-mode v-if="darkMode" />
        <i-ic-round-light-mode v-else />
      </NavbarIcon>
      <!-- Admin Settings -->
      <NavbarIcon
        v-if="user?.admin"
        class="navbar-icon navbar-clickable"
        :title="$t('admin.settings.settings')"
        :to="{ name: 'admin-settings' }"
      >
        <i-clarity-settings-solid />
      </NavbarIcon>

      <!-- Active Builds Indicator -->
      <ActiveBuilds v-if="user" class="navbar-icon navbar-clickable" />

      <!-- User Avatar -->
      <NavbarIcon
        v-if="user"
        :to="{ name: 'user' }"
        :title="$t('user.settings')"
        class="navbar-icon navbar-clickable !p-1.5"
      >
        <img v-if="user && user.avatar_url" class="rounded-full" :src="`${user.avatar_url}`" />
      </NavbarIcon>
      <!-- Login Button -->
      <Button v-else :text="$t('login')" @click="doLogin" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { useRoute } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import { useDarkMode } from '~/compositions/useDarkMode';

import ActiveBuilds from './ActiveBuilds.vue';
import NavbarIcon from './NavbarIcon.vue';

export default defineComponent({
  name: 'Navbar',

  components: { Button, ActiveBuilds, NavbarIcon },

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

<style scoped>
.navbar-link {
  @apply px-3 py-2 -my-1 rounded-md;
}

.navbar-clickable {
  @apply hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 transition-colors duration-100;
}
</style>
