<template>
  <main class="flex h-full w-full flex-col items-center justify-center">
    <Error v-if="errorMessage" class="md:w-3xl w-full">
      <span class="whitespace-pre">{{ errorMessage }}</span>
      <span v-if="errorDescription" class="mt-1 whitespace-pre">{{ errorDescription }}</span>
      <a
        v-if="errorUri"
        :href="errorUri"
        target="_blank"
        class="text-wp-link-100 hover:text-wp-link-200 mt-1 cursor-pointer"
      >
        <span>{{ errorUri }}</span>
      </a>
    </Error>

    <div
      class="border-wp-background-400 bg-wp-background-100 dark:bg-wp-background-200 md:w-3xl md:h-sm flex w-full flex-col overflow-hidden border shadow md:m-8 md:flex-row md:rounded-md"
    >
      <div class="bg-wp-primary-200 dark:bg-wp-primary-300 flex min-h-48 items-center justify-center md:w-3/5">
        <WoodpeckerLogo preserveAspectRatio="xMinYMin slice" class="w-30 h-30 md:h-48 md:w-48" />
      </div>
      <div class="flex min-h-48 flex-col items-center justify-center gap-4 text-center md:w-2/5">
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
