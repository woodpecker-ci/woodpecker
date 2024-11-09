<template>
  <Settings :title="$t('user.settings.cli_and_api.cli_and_api')" :desc="$t('user.settings.cli_and_api.desc')">
    <InputField :label="$t('user.settings.cli_and_api.cli_usage')">
      <template #titleActions>
        <a :href="cliDownload" target="_blank" class="ml-4 text-wp-link-100 hover:text-wp-link-200">{{
          $t('user.settings.cli_and_api.download_cli')
        }}</a>
      </template>
      <pre class="code-box">{{ usageWithCli }}</pre>
    </InputField>

    <InputField :label="$t('user.settings.cli_and_api.token')">
      <template #titleActions>
        <Button class="ml-auto" :text="$t('user.settings.cli_and_api.reset_token')" @click="resetToken" />
      </template>
      <pre class="code-box">{{ token }}</pre>
    </InputField>

    <InputField :label="$t('user.settings.cli_and_api.api_usage')">
      <template #titleActions>
        <a
          v-if="enableSwagger"
          :href="`${address}/swagger/index.html`"
          target="_blank"
          class="ml-4 text-wp-link-100 hover:text-wp-link-200"
        >
          {{ $t('user.settings.cli_and_api.swagger_ui') }}
        </a>
      </template>
      <pre class="code-box">{{ usageWithCurl }}</pre>
    </InputField>
  </Settings>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import InputField from '~/components/form/InputField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import useConfig from '~/compositions/useConfig';

const { rootPath, enableSwagger } = useConfig();

const apiClient = useApiClient();
const token = ref<string | undefined>();

onMounted(async () => {
  token.value = await apiClient.getToken();
});

const address = `${window.location.protocol}//${window.location.host}${rootPath}`; // port is included in location.host

const usageWithCurl = computed(() => {
  let usage = `export WOODPECKER_SERVER="${address}"\n`;
  usage += `export WOODPECKER_TOKEN="${token.value}"\n`;
  usage += `\n`;
  usage += `# curl -i \${WOODPECKER_SERVER}/api/user -H "Authorization: Bearer \${WOODPECKER_TOKEN}"`;
  return usage;
});

const usageWithCli = `# woodpecker setup --server ${address}`;

const cliDownload = 'https://github.com/woodpecker-ci/woodpecker/releases';

const resetToken = async () => {
  token.value = await apiClient.resetToken();
  window.location.href = `${address}/logout`;
};
</script>
