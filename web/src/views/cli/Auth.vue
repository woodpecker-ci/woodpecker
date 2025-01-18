<template>
  <div class="m-auto flex flex-col gap-4">
    <div class="text-center text-wp-text-100">
      <WoodpeckerLogo preserveAspectRatio="xMinYMin slice" class="m-auto mb-8 w-32" />
      <template v-if="state === 'confirm'">
        <h1 class="text-4xl font-bold">{{ $t('login_to_cli') }}</h1>
        <p class="text-2xl">{{ $t('login_to_cli_description') }}</p>
      </template>
      <template v-else-if="state === 'success'">
        <h1 class="text-4xl font-bold">{{ $t('cli_login_success') }}</h1>
        <p class="text-2xl">{{ $t('return_to_cli') }}</p>
      </template>
      <template v-else-if="state === 'failed'">
        <h1 class="mt-4 text-4xl font-bold">{{ $t('cli_login_failed') }}</h1>
        <p class="text-2xl">{{ $t('return_to_cli') }}</p>
      </template>
      <template v-else-if="state === 'denied'">
        <h1 class="mt-4 text-4xl font-bold">{{ $t('cli_login_denied') }}</h1>
        <p class="text-2xl">{{ $t('return_to_cli') }}</p>
      </template>
    </div>

    <div v-if="state === 'confirm'" class="flex justify-center gap-4">
      <Button :text="$t('login_to_cli')" color="green" @click="sendToken(false)" />
      <Button :text="$t('abort')" color="red" @click="abortLogin" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import WoodpeckerLogo from '~/assets/logo.svg?component';
import Button from '~/components/atomic/Button.vue';
import useApiClient from '~/compositions/useApiClient';

const apiClient = useApiClient();
const route = useRoute();
const { t } = useI18n();
const state = ref<'confirm' | 'success' | 'failed' | 'denied'>('confirm');

async function sendToken(abort = false) {
  const port = route.query.port as string;
  if (!port) {
    throw new Error('Unexpected: port not found');
  }

  const address = `http://localhost:${port}`;

  const token = abort ? '' : await apiClient.getToken();

  const resp = await fetch(`${address}/token`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ token }),
  });

  if (abort) {
    state.value = 'denied';
    window.close();
    return;
  }

  const data = (await resp.json()) as { ok: string };
  if (data.ok === 'true') {
    state.value = 'success';
  } else {
    state.value = 'failed';
    // eslint-disable-next-line no-alert
    alert(t('cli_login_failed'));
  }
}

async function abortLogin() {
  await sendToken(true);
}
</script>
