<template>
  <main class="flex flex-col w-full h-full justify-center items-center">
    <!-- TODO: Should use vue notifications. -->
    <Error v-if="errorMessage" text-only :text="errorMessage" class="w-full md:w-3xl" />

    <div
      class="flex flex-col w-full overflow-hidden bg-wp-background-100 shadow border border-wp-background-400 dark:bg-wp-background-200 md:m-8 md:rounded-md md:flex-row md:w-3xl md:h-sm"
    >
      <div class="flex justify-center items-center bg-wp-primary-200 dark:bg-wp-primary-300 min-h-48 md:w-3/5">
        <WoodpeckerLogo preserveAspectRatio="xMinYMin slice" class="w-30 h-30 md:w-48 md:h-48" />
      </div>
      <div class="flex justify-center items-center flex-col md:w-2/5 min-h-48 gap-4 text-center">
        <h1 class="text-xl text-wp-text-100">{{ $t('welcome') }}</h1>
        <Button @click="doLogin">{{ $t('login') }}</Button>
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
import useAuthentication from '~/compositions/useAuthentication';

const route = useRoute();
const router = useRouter();
const authentication = useAuthentication();
const errorMessage = ref<string>();
const i18n = useI18n();

function doLogin() {
  const url = typeof route.query.url === 'string' ? route.query.url : '';
  authentication.authenticate(url);
}

const authErrorMessages = {
  oauth_error: i18n.t('user.oauth_error'),
  internal_error: i18n.t('user.internal_error'),
  access_denied: i18n.t('user.access_denied'),
};

onMounted(async () => {
  if (authentication.isAuthenticated) {
    await router.replace({ name: 'home' });
    return;
  }

  if (route.query.code) {
    const code = route.query.code as keyof typeof authErrorMessages;
    errorMessage.value = authErrorMessages[code];
  }
});
</script>
