<template>
  <nav
    class="text-neutral-content flex border-b border-wp-background-100 bg-wp-primary-200 p-4 font-bold text-wp-primary-text-100 dark:bg-wp-primary-300"
  >
    <div class="flex items-center space-x-2">
      <router-link :to="{ name: 'home' }" class="-my-2 flex flex-col px-2">
        <WoodpeckerLogo class="h-8 w-8" />
        <span class="text-xs" :title="version?.current">{{ version?.currentShort }}</span>
      </router-link>
      <router-link v-if="user" :to="{ name: 'repos' }" class="navbar-clickable navbar-link">
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="hidden md:flex">{{ $t('repositories') }}</span>
      </router-link>
      <a href="https://woodpecker-ci.org/" target="_blank" class="navbar-clickable navbar-link hidden md:flex">{{
        $t('docs')
      }}</a>
      <a v-if="enableSwagger" :href="apiUrl" target="_blank" class="navbar-clickable navbar-link hidden md:flex">{{
        $t('api')
      }}</a>
    </div>
    <div class="-m-1.5 ml-auto flex items-center space-x-2">
      <IconButton
        v-if="user?.admin"
        class="navbar-icon relative"
        :title="$t('settings')"
        :to="{ name: 'admin-settings' }"
      >
        <Icon name="settings" />
        <div v-if="version?.needsUpdate" class="absolute right-2 top-2 h-3 w-3 rounded-full bg-int-wp-error-100" />
      </IconButton>

      <ActivePipelines v-if="user" class="navbar-icon !p-1.5" />
      <IconButton v-if="user" :to="{ name: 'user' }" :title="$t('user.settings.settings')" class="navbar-icon !p-1.5">
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
  @apply h-11 w-11 rounded-md p-2.5;
}

.navbar-icon :deep(svg) {
  @apply h-full w-full;
}

.navbar-link {
  @apply hover-effect -my-1 rounded-md px-3 py-2;
}
</style>
