<template>
  <div class="flex flex-col gap-4 m-auto">
    <div class="text-center text-wp-text-100">
      <h1 class="text-2xl font-bold">{{ $t('login_to_cli') }}</h1>
      <p>{{ $t('login_to_cli_description') }}</p>
    </div>

    <div class="flex gap-4 justify-center">
      <Button :text="$t('login_to_cli')" color="green" @click="sendToken" />
      <Button :text="$t('abort')" color="red" @click="abortLogin" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import useApiClient from '~/compositions/useApiClient';

const apiClient = useApiClient();
const route = useRoute();
const { t } = useI18n();

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
    window.close();
    return;
  }

  const data = (await resp.json()) as { ok: string };
  if (data.ok === 'true') {
    window.close();
  } else {
    // eslint-disable-next-line no-alert
    alert(t('cli_login_failed'));
  }
}

async function abortLogin() {
  await sendToken(true);
}
</script>
