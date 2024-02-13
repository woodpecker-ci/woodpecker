<template>
  <div class="flex gap-4 m-auto">
    <Button :text="$t('login_to_cli')" @click="sendToken" />
    <Button :text="$t('abort_cli_login')" @click="abortLogin" />
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

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
