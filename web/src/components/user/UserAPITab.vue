<template>
  <Settings :title="$t('user.settings.api.api')" :desc="$t('user.settings.api.desc')">
    <div class="mt-2 mb-4">
      <div class="flex items-center mb-2">
        <div class="flex flex-row items-center text-wp-text-100 font-bold">
          <label>{{ $t('user.settings.api.token') }}</label>
        </div>
        <Button class="ml-auto" :text="$t('user.settings.api.reset_token')" @click="resetToken" />
      </div>
      <pre class="code-box">{{ token }}</pre>
    </div>

    <div class="mt-2 mb-4">
      <div class="flex items-center text-wp-text-100 font-bold mb-2">
        <label>{{ $t('user.settings.api.shell_setup') }}</label>
      </div>
      <pre class="code-box">{{ usageWithShell }}</pre>
    </div>

    <div class="mt-2 mb-4">
      <div class="flex items-center mb-2">
        <div class="flex items-center text-wp-text-100 font-bold">
          <label>{{ $t('user.settings.api.api_usage') }}</label>
        </div>
        <a
          v-if="enableSwagger"
          :href="`${address}/swagger/index.html`"
          target="_blank"
          class="ml-4 text-wp-link-100 hover:text-wp-link-200"
          >{{ $t('user.settings.api.swagger_ui') }}</a
        >
      </div>
      <pre class="code-box">{{ usageWithCurl }}</pre>
    </div>

    <div class="mt-2 mb-4">
      <div class="flex items-center mb-2">
        <div class="flex items-center text-wp-text-100 font-bold">
          <label>{{ $t('user.settings.api.cli_usage') }}</label>
        </div>
        <a :href="cliDownload" target="_blank" class="ml-4 text-wp-link-100 hover:text-wp-link-200">{{
          $t('user.settings.api.dl_cli')
        }}</a>
      </div>
      <pre class="code-box">{{ usageWithCli }}</pre>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import useConfig from '~/compositions/useConfig';

const { t } = useI18n();
const { rootPath, enableSwagger } = useConfig();

const apiClient = useApiClient();
const token = ref<string | undefined>();

onMounted(async () => {
  token.value = await apiClient.getToken();
});

const address = `${window.location.protocol}//${window.location.host}${rootPath}`; // port is included in location.host

const usageWithShell = computed(() => {
  let usage = `export WOODPECKER_SERVER="${address}"\n`;
  usage += `export WOODPECKER_TOKEN="${token.value}"\n`;
  return usage;
});

const usageWithCurl = `# ${t(
  'user.settings.api.shell_setup_before',
)}\ncurl -i \${WOODPECKER_SERVER}/api/user -H "Authorization: Bearer \${WOODPECKER_TOKEN}"`;

const usageWithCli = `# ${t('user.settings.api.shell_setup_before')}\nwoodpecker info`;

const cliDownload = 'https://github.com/woodpecker-ci/woodpecker/releases';

const resetToken = async () => {
  token.value = await apiClient.resetToken();
  window.location.href = `${address}/logout`;
};
</script>
