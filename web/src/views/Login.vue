<template>
  <div class="flex flex-col w-full h-full justify-center items-center">
    <div v-if="errorMessage" class="bg-red-400 text-white dark:text-gray-500 p-4 rounded-md text-lg">
      {{ errorMessage }}
    </div>

    <div
      class="
        flex flex-col
        w-full
        overflow-hidden
        md:m-8 md:rounded-md md:shadow md:border md:bg-white md:dark:bg-dark-gray-700
        dark:border-dark-200
        md:flex-row md:w-3xl md:h-sm
        justify-center
      "
    >
      <div class="flex md:bg-lime-500 md:dark:bg-lime-900 md:w-3/5 justify-center items-center">
        <img class="w-48 h-48" src="../assets/logo.svg?url" />
      </div>
      <div class="flex flex-col my-8 md:w-2/5 p-4 items-center justify-center">
        <h1 class="text-xl text-gray-600 dark:text-gray-500">{{ $t('welcome') }}</h1>
        <Button class="mt-4" @click="doLogin">{{ $t('login') }}</Button>
      </div>
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import useAuthentication from '~/compositions/useAuthentication';

export default defineComponent({
  name: 'Login',

  components: {
    Button,
  },

  setup() {
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

    return {
      doLogin,
      errorMessage,
    };
  },
});
</script>
