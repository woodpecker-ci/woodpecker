<template>
  <nav
    class="flex border-wp-background-100 bg-wp-primary-200 dark:bg-wp-primary-300 p-4 border-b font-bold text-neutral-content text-wp-primary-text-100"
  >
    <div class="flex items-center space-x-2">
      <router-link :to="{ name: 'home' }" class="flex flex-col -my-2 px-2">
        <WoodpeckerLogo class="w-8 h-8" />
        <span class="text-xs" :title="version?.current">{{ version?.currentShort }}</span>
      </router-link>
      <router-link v-if="user" :to="{ name: 'repos' }" class="navbar-clickable navbar-link">
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="md:flex hidden">{{ $t('repositories') }}</span>
      </router-link>
      <a href="https://woodpecker-ci.org/" target="_blank" class="md:flex hidden navbar-clickable navbar-link">{{
        $t('docs')
      }}</a>
      <a v-if="enableSwagger" :href="apiUrl" target="_blank" class="md:flex hidden navbar-clickable navbar-link">{{
        $t('api')
      }}</a>
    </div>
    <div class="flex items-center space-x-2 -m-1.5 ml-auto">
      <IconButton
        v-if="user?.admin"
        class="relative navbar-icon"
        :title="$t('settings')"
        :to="{ name: 'admin-settings' }"
      >
        <span v-if="user?.admin" class="text-xs" :title="$t('admin')">{{ $t('admin') }}</span>
        <Icon name="settings" />
        <div
          v-if="version?.needsUpdate"
          class="top-2 right-2 absolute bg-int-wp-state-error-100 rounded-full w-3 h-3"
        />
      </IconButton>

      <ActivePipelines v-if="user" class="!p-1.5 navbar-icon" />
      <IconButton v-if="user" :to="{ name: 'user' }" :title="$t('user.settings.settings')" class="navbar-icon">
        <img v-if="user && user.avatar_url" class="rounded-md" :src="`${user.avatar_url}`" />
      </IconButton>
      <Button v-else :text="$t('login')" :to="`/login?url=${route.fullPath}`" />
    </div>
  </nav>
</template>

<script lang="ts" setup>
import { useRoute } from 'vue-router';

import WoodpeckerLogo from '~/assets/logo.svg?component';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
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

const { enableSwagger } = config;
</script>

<style scoped>
.navbar-icon {
  @apply p-2.5 rounded-md w-11 h-11;
}

.navbar-icon :deep(svg) {
  @apply w-full h-full;
}

.navbar-link {
  @apply -my-1 px-3 py-2 rounded-md hover-effect;
}
</style>
