<template>
  <!-- Navbar -->
  <nav
    class="flex bg-wp-primary-400 text-neutral-content p-4 dark:bg-wp-darkgray-800 dark:border-b dark:border-wp-gray-700 font-bold"
  >
    <!-- Left Links Box -->
    <div class="flex text-white dark:text-wp-gray-500 items-center space-x-2">
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
      <!-- API Link -->
      <a :href="apiUrl" target="_blank" class="navbar-link navbar-clickable hidden md:flex">{{ $t('api') }}</a>
    </div>
    <!-- Right Icons Box -->
    <div class="flex ml-auto -m-1.5 items-center space-x-2 text-white dark:text-wp-gray-400">
      <!-- Dark Mode Toggle -->
      <IconButton
        :icon="darkMode ? 'dark' : 'light'"
        :title="$t(darkMode ? 'color_scheme_dark' : 'color_scheme_light')"
        class="navbar-icon"
        @click="darkMode = !darkMode"
      />
      <!-- Admin Settings -->
      <IconButton
        v-if="user?.admin"
        class="navbar-icon"
        :title="$t('admin.settings.settings')"
        :to="{ name: 'admin-settings' }"
      >
        <i-clarity-settings-solid />
      </IconButton>

      <!-- Active Pipelines Indicator -->
      <ActivePipelines v-if="user" class="navbar-icon" />
      <!-- User Avatar -->
      <IconButton v-if="user" :to="{ name: 'user' }" :title="$t('user.settings')" class="navbar-icon !p-1.5">
        <img v-if="user && user.avatar_url" class="rounded-md" :src="`${user.avatar_url}`" />
      </IconButton>
      <!-- Login Button -->
      <Button v-else :text="$t('login')" @click="doLogin" />
    </div>
  </nav>
</template>

<script lang="ts">
import { defineComponent } from 'vue';
import { useRoute } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import { useDarkMode } from '~/compositions/useDarkMode';

import ActivePipelines from './ActivePipelines.vue';

export default defineComponent({
  name: 'Navbar',

  components: { Button, ActivePipelines, IconButton },

  setup() {
    const config = useConfig();
    const route = useRoute();
    const authentication = useAuthentication();
    const { darkMode } = useDarkMode();
    const docsUrl = config.docs || undefined;
    const apiUrl = `${config.rootURL ?? ''}/swagger/index.html`;

    function doLogin() {
      authentication.authenticate(route.fullPath);
    }

    const version = config.version?.startsWith('next') ? 'next' : config.version;

    return { darkMode, user: authentication.user, doLogin, docsUrl, version, apiUrl };
  },
});
</script>

<style scoped>
.navbar-icon {
  @apply w-11 h-11 rounded-md p-2.5;
}

.navbar-icon :deep(svg) {
  @apply w-full h-full;
}

.navbar-link {
  @apply px-3 py-2 -my-1 rounded-md hover-effect;
}
</style>
