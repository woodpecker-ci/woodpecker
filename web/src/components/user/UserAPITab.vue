<template>
  <Panel>
    <div>
      <div class="flex items-center mb-2">
        <h2 class="text-lg text-wp-text-100">{{ $t('user.api.token') }}</h2>
        <Button class="ml-4" :text="$t('user.api.reset_token')" @click="resetToken" />
      </div>
      <pre class="code-box">{{ token }}</pre>
    </div>

    <div>
      <h2 class="text-lg text-wp-text-100">{{ $t('user.api.shell_setup') }}</h2>
      <pre class="code-box">{{ usageWithShell }}</pre>
    </div>

    <div>
      <div class="flex items-center">
        <h2 class="text-lg text-wp-text-100">{{ $t('user.api.api_usage') }}</h2>
        <a
          :href="`${address}/swagger/index.html`"
          target="_blank"
          class="ml-4 text-wp-link-100 hover:text-wp-link-200"
          >{{ $t('user.api.swagger_ui') }}</a
        >
      </div>
      <pre class="code-box">{{ usageWithCurl }}</pre>
    </div>

    <div>
      <div class="flex items-center">
        <h2 class="text-lg text-wp-text-100">{{ $t('user.api.cli_usage') }}</h2>
        <a :href="cliDownload" target="_blank" class="ml-4 text-wp-link-100 hover:text-wp-link-200">{{
          $t('user.api.dl_cli')
        }}</a>
      </div>
      <pre class="code-box">{{ usageWithCli }}</pre>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { setI18nLanguage } from '~/compositions/useI18n';

const { t, locale } = useI18n();

const apiClient = useApiClient();
const token = ref<string | undefined>();

onMounted(async () => {
  token.value = await apiClient.getToken();
});

// eslint-disable-next-line no-restricted-globals
const address = `${location.protocol}//${location.host}`; // port is included in location.host

const usageWithShell = computed(() => {
  let usage = `export WOODPECKER_SERVER="${address}"\n`;
  usage += `export WOODPECKER_TOKEN="${token.value}"\n`;
  return usage;
});

const usageWithCurl = `# ${t(
  'user.api.shell_setup_before',
)}\ncurl -i \${WOODPECKER_SERVER}/api/user -H "Authorization: Bearer \${WOODPECKER_TOKEN}"`;

const usageWithCli = `# ${t('user.api.shell_setup_before')}\nwoodpecker info`;

const cliDownload = 'https://github.com/woodpecker-ci/woodpecker/releases';

const resetToken = async () => {
  token.value = await apiClient.resetToken();
  window.location.href = `${address}/logout`;
};
</script>
