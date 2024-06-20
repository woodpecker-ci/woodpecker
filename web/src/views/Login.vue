<template>
  <main class="flex flex-col w-full h-full justify-center items-center">
    <Error v-if="errorMessage" :text="errorMessage" class="w-full md:w-3xl" />

    <div
      class="flex flex-col w-full overflow-hidden bg-wp-background-100 shadow border border-wp-background-400 dark:bg-wp-background-200 md:m-8 md:rounded-md md:flex-row md:w-3xl md:h-sm"
    >
      <div class="flex justify-center items-center bg-wp-primary-200 dark:bg-wp-primary-300 min-h-48 md:w-3/5">
        <WoodpeckerLogo preserveAspectRatio="xMinYMin slice" class="w-30 h-30 md:w-48 md:h-48" />
      </div>
      <div class="flex justify-center items-center flex-col md:w-2/5 min-h-48 gap-4 text-center">
        <h1 class="text-xl text-wp-text-100">{{ $t('welcome') }}</h1>
        <div class="flex flex-col gap-2">
          <Button v-for="forge in forges" :key="forge.id" @click="doLogin(forge.id)">{{
            $t('login_with', { forge: getHostFromUrl(forge) })
          }}</Button>
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
import useAuthentication from '~/compositions/useAuthentication';

const route = useRoute();
const router = useRouter();
const authentication = useAuthentication();
const errorMessage = ref<string>();
const i18n = useI18n();

type Forge = {
  id: number;
  url: string;
  type: string;
};

const forges = ref<Forge[]>([
  {
    id: 1,
    url: 'http://localhost:3000/',
    type: 'gitea',
  },
  {
    id: 2,
    url: '',
    type: 'github',
  },
]);

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
};

onMounted(async () => {
  if (authentication.isAuthenticated) {
    await router.replace({ name: 'home' });
    return;
  }

  if (route.query.error) {
    const error = route.query.error as keyof typeof authErrorMessages;
    errorMessage.value = authErrorMessages[error] ?? error;

    if (route.query.error_description) {
      errorMessage.value += `\n${route.query.error_description}`;
    }

    if (route.query.error_uri) {
      errorMessage.value += `\n${route.query.error_uri}`;
    }
  }
});
</script>
