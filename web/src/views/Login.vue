<template>
  <main class="flex flex-col justify-center items-center w-full h-full">
    <Error v-if="errorMessage" class="w-full md:w-3xl">
      <span class="whitespace-pre">{{ errorMessage }}</span>
      <span v-if="errorDescription" class="mt-1 whitespace-pre">{{ errorDescription }}</span>
      <a
        v-if="errorUri"
        :href="errorUri"
        target="_blank"
        class="mt-1 text-wp-link-100 hover:text-wp-link-200 cursor-pointer"
      >
        <span>{{ errorUri }}</span>
      </a>
    </Error>

    <div
      class="flex md:flex-row flex-col border-wp-background-400 bg-wp-background-100 dark:bg-wp-background-200 shadow md:m-8 border md:rounded-md w-full md:w-3xl md:h-sm overflow-hidden"
    >
      <div class="flex justify-center items-center bg-wp-primary-200 dark:bg-wp-primary-300 md:w-3/5 min-h-48">
        <WoodpeckerLogo preserveAspectRatio="xMinYMin slice" class="w-30 md:w-48 h-30 md:h-48" />
      </div>
      <div class="flex flex-col justify-center items-center gap-4 md:w-2/5 min-h-48 text-center">
        <h1 class="text-wp-text-100 text-xl">{{ $t('welcome') }}</h1>
        <div class="flex flex-col gap-2">
          <Button
            v-for="forge in forges"
            :key="forge.id"
            :start-icon="forge.type === 'addon' ? 'repo' : forge.type"
            @click="doLogin(forge.id)"
          >
            {{ $t('login_with', { forge: getHostFromUrl(forge) }) }}
          </Button>
        </div>
      </div>
    </div>
  </main>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import WoodpeckerLogo from '~/assets/logo.svg?component';
import Button from '~/components/atomic/Button.vue';
import Error from '~/components/atomic/Error.vue';
import useApiClient from '~/compositions/useApiClient';
import useAuthentication from '~/compositions/useAuthentication';
import type { Forge } from '~/lib/api/types';

const route = useRoute();
const router = useRouter();
const authentication = useAuthentication();
const i18n = useI18n();
const apiClient = useApiClient();

const forges = ref<Forge[]>([]);

function getHostFromUrl(forge: Forge) {
  if (!forge.url) {
    return forge.type.charAt(0).toUpperCase() + forge.type.slice(1);
  }

  const url = new URL(forge.url);
  return url.hostname;
}

function doLogin(forgeId?: number) {
  const url = typeof route.query.url === 'string' ? route.query.url : '';
  authentication.authenticate(url, forgeId);
}

const authErrorMessages = {
  oauth_error: i18n.t('oauth_error'),
  internal_error: i18n.t('internal_error'),
  registration_closed: i18n.t('registration_closed'),
  access_denied: i18n.t('access_denied'),
  invalid_state: i18n.t('invalid_state'),
};

const errorMessage = ref<string>();
const errorDescription = ref<string>(route.query.error_description as string);
const errorUri = ref<string>(route.query.error_uri as string);

onMounted(async () => {
  if (authentication.isAuthenticated) {
    await router.replace({ name: 'home' });
    return;
  }

  forges.value = (await apiClient.getForges()) ?? [];

  if (route.query.error) {
    const error = route.query.error as keyof typeof authErrorMessages;
    errorMessage.value = authErrorMessages[error] ?? error;
  }
});
</script>
