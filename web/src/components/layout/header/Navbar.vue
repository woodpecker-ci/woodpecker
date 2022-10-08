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
      <router-link v-if="user" :to="{ name: 'repos' }" class="navbar-link">
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="hidden md:flex">{{ $t('repositories') }}</span>
      </router-link>
      <!-- Docs Link -->
      <a :href="docsUrl" target="_blank" class="navbar-link hidden md:flex">{{ $t('docs') }}</a>
    </div>
    <!-- Right Icons Box -->
    <div class="flex ml-auto items-center space-x-3 text-white dark:text-gray-400">
      <!-- Dark Mode Toggle -->
      <IconButton
        :icon="darkMode ? 'dark' : 'light'"
        class="!text-white !dark:text-gray-500 navbar-icon"
        :title="darkMode ? $t('color_scheme_dark') : $t('color_scheme_light')"
        @click="darkMode = !darkMode"
      />
      <!-- Admin Settings -->
      <IconButton
        v-if="user?.admin"
        icon="settings"
        class="!text-white !dark:text-gray-500 navbar-icon"
        :title="$t('admin.settings.settings')"
        :to="{ name: 'admin-settings' }"
      />
      <!-- Active Builds Indicator -->
      <ActiveBuilds v-if="user" />
      <!-- User Avatar -->
      <router-link v-if="user" :to="{ name: 'user' }" class="rounded-full overflow-hidden">
        <img v-if="user && user.avatar_url" class="navbar-icon" :src="`${user.avatar_url}`" />
      </router-link>
      <!-- Login Button -->
      <Button v-else :text="$t('login')" @click="doLogin" />
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

<style scoped>
.navbar-link {
  @apply hover:bg-black hover:bg-opacity-10 transition-colors duration-100 px-3 py-2 -my-1 rounded-md;
}

.navbar-icon {
  @apply w-8 h-8;
}
</style>
