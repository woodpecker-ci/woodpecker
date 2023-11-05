<template>
  <nav
    class="flex bg-wp-primary-200 dark:bg-wp-primary-300 text-neutral-content p-4 border-b border-wp-background-100 font-bold text-wp-primary-text-100"
  >
    <div class="flex items-center space-x-2">
      <router-link :to="{ name: 'home' }" class="flex flex-col -my-2 px-2">
        <WoodpeckerLogo class="w-8 h-8" />
        <span class="text-xs" :title="version?.current">{{ version?.currentShort }}</span>
      </router-link>
      <router-link v-if="user" :to="{ name: 'repos' }" class="navbar-link navbar-clickable">
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="hidden md:flex">{{ $t('repositories') }}</span>
      </router-link>
      <a href="https://woodpecker-ci.org/" target="_blank" class="navbar-link navbar-clickable hidden md:flex">{{
        $t('docs')
      }}</a>
      <a v-if="enableSwagger" :href="apiUrl" target="_blank" class="navbar-link navbar-clickable hidden md:flex">{{
        $t('api')
      }}</a>
    </div>
    <div class="flex ml-auto -m-1.5 items-center space-x-2">
      <div v-if="user?.admin" class="relative">
        <IconButton
          class="navbar-icon"
          :title="$t('admin.settings.settings')"
          :to="{ name: 'admin-settings' }"
          icon="settings"
        />
        <div
          v-if="version?.needsUpdate"
          class="absolute top-2 right-2 bg-int-wp-state-error-100 rounded-full w-3 h-3"
        />
      </div>

      <ActivePipelines v-if="user" class="navbar-icon" />
      <IconButton v-if="user" :to="{ name: 'user' }" :title="$t('user.settings.settings')" class="navbar-icon !p-1.5">
        <img v-if="user && user.avatar_url" class="rounded-md" :src="`${user.avatar_url}`" />
      </IconButton>
      <Button v-else :text="$t('login')" @click="doLogin" />
    </div>
  </nav>
</template>

<script lang="ts" setup>
import { useRoute } from 'vue-router';

import WoodpeckerLogo from '~/assets/logo.svg?component';
import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import { useVersion } from '~/compositions/useVersion';

import ActivePipelines from './ActivePipelines.vue';

const version = useVersion();
const config = useConfig();
const route = useRoute();
const authentication = useAuthentication();
const { user } = authentication;
const apiUrl = `${config.rootPath ?? ''}/swagger/index.html`;

function doLogin() {
  authentication.authenticate(route.fullPath);
}

const { enableSwagger } = config;
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
