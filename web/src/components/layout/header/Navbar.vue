<template>
  <nav
    class="text-neutral-content bg-wp-primary-200 text-wp-primary-text-100 dark:bg-wp-primary-300 flex border-b border-black/20 p-4 font-bold"
  >
    <div class="flex items-center space-x-2">
      <router-link :to="{ name: 'home' }" class="-my-2 flex flex-col px-2">
        <WoodpeckerLogo class="h-8 w-8" />
        <span class="text-xs" :title="version?.current">{{ version?.currentShort }}</span>
      </router-link>
      <router-link v-if="user" :to="{ name: 'repos' }" class="navbar-clickable navbar-link">
        <span class="flex md:hidden">{{ $t('repos') }}</span>
        <span class="hidden md:flex">{{ $t('repositories.title') }}</span>
      </router-link>
      <a href="https://woodpecker-ci.org/" target="_blank" class="navbar-clickable navbar-link hidden md:flex">{{
        $t('docs')
      }}</a>
      <a v-if="enableSwagger" :href="apiUrl" target="_blank" class="navbar-clickable navbar-link hidden md:flex">{{
        $t('api')
      }}</a>
    </div>
    <div class="-m-1.5 ml-auto flex items-center space-x-2">
      <ViewLevelSelector class="hidden md:flex mr-1" />

      <!-- SignalBox Clinical Assurance & Governance Toggle -->
      <button
        data-testid="signalbox-toggle"
        class="signalbox-toggle-btn navbar-icon relative flex items-center justify-center text-cyan-200 hover:text-cyan-100"
        title="Symphonix SignalBox Clinical Assurance"
        @click="signalBox.toggle"
      >
        <svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path stroke-linecap="round" stroke-linejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
        </svg>
        <span class="absolute top-1.5 right-1.5 h-2 w-2 rounded-full bg-cyan-400 animate-pulse" />
      </button>

      <IconButton
        v-if="user?.admin"
        class="navbar-icon relative"
        :title="$t('settings')"
        :to="{ name: 'admin-settings' }"
      >
        <Icon name="settings" />
        <div v-if="version?.needsUpdate" class="bg-wp-error-100 absolute top-2 right-2 h-3 w-3 rounded-full" />
      </IconButton>

      <ActivePipelines v-if="user" class="navbar-icon p-1.5!" />
      <IconButton v-if="user" :to="{ name: 'user' }" :title="$t('user.settings.settings')" class="navbar-icon p-1.5!">
        <img v-if="user && user.avatar_url" class="rounded-md" :src="`${user.avatar_url}`" />
      </IconButton>
      <Button
        v-else
        :text="$t('login')"
        :to="{ name: 'login' }"
        class="navbar-link !text-wp-primary-text-100 bg-wp-primary-200 dark:bg-wp-primary-300 !border-transparent"
        @click="saveRedirect"
      />
    </div>
  </nav>
</template>

<script lang="ts" setup>
import { useRoute } from 'vue-router';

import WoodpeckerLogo from '~/assets/logo.svg?component';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ViewLevelSelector from '~/components/symphonix/ViewLevelSelector.vue';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import useSignalBox from '~/compositions/useSignalBox';
import useUserConfig from '~/compositions/useUserConfig';
import { useVersion } from '~/compositions/useVersion';

import ActivePipelines from './ActivePipelines.vue';

const signalBox = useSignalBox();

const version = useVersion();
const config = useConfig();
const userConfig = useUserConfig();
const route = useRoute();
const authentication = useAuthentication();
const { user } = authentication;
const apiUrl = `${config.rootPath ?? ''}/swagger/index.html`;

const { enableSwagger } = config;

function saveRedirect() {
  userConfig.setUserConfig('redirectUrl', route.fullPath);
}
</script>

<style scoped>
@reference '~/tailwind.css';

.navbar-icon {
  @apply h-11 w-11 rounded-md p-2.5 hover:bg-black/20! dark:hover:bg-white/10!;
}

.navbar-icon :deep(svg) {
  @apply h-full w-full;
}

.navbar-link {
  @apply -my-1 rounded-md px-3 py-2 hover:bg-black/20! dark:hover:bg-white/10!;
}
</style>
