<template>
  <!-- Navbar -->
  <div class="flex bg-lime-600 text-neutral-content p-4 dark:bg-dark-gray-800 dark:border-b dark:border-gray-700">
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
      <IconButton
        :icon="darkMode ? 'dark' : 'light'"
        :title="$t(darkMode ? 'color_scheme_dark' : 'color_scheme_light')"
        class="navbar-icon navbar-clickable"
        @click="darkMode = !darkMode"
      />
      <!-- Admin Settings -->
      <IconButton
        v-if="user?.admin"
        class="navbar-icon navbar-clickable"
        :title="$t('admin.settings.settings')"
        :to="{ name: 'admin-settings' }"
      >
        <i-clarity-settings-solid />
      </IconButton>

      <!-- Active Pipelines Indicator -->
      <ActivePipelines v-if="user" class="navbar-icon navbar-clickable" />
      <!-- User Avatar -->
      <IconButton
        v-if="user"
        :to="{ name: 'user' }"
        :title="$t('user.settings')"
        class="navbar-icon navbar-clickable !p-1.5"
      >
        <img v-if="user && user.avatar_url" class="rounded-md" :src="`${user.avatar_url}`" />
      </IconButton>
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

import ActivePipelines from './ActivePipelines.vue';

export default defineComponent({
  name: 'Navbar',

  components: { Button, ActivePipelines, IconButton },

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
.navbar-icon {
  @apply w-11 h-11 rounded-full p-2.5;
}

.navbar-icon :deep(svg) {
  @apply w-full h-full;
}

.navbar-link {
  @apply px-3 py-2 -my-1 rounded-md;
}

.navbar-clickable {
  @apply hover:bg-black hover:bg-opacity-10 dark:hover:bg-white dark:hover:bg-opacity-5 transition-colors duration-100;
}
</style>
