<template>
  <Scaffold>
    <template #title>{{ $t('user.settings') }}</template>
    <template #titleActions><Button :text="$t('logout')" :to="`${address}/logout`" /></template>
    <div class="space-y-4 flex flex-col">
      <div>
        <h2 class="text-lg text-color">{{ $t('user.token') }}</h2>
        <pre class="cli-box">{{ token }}</pre>
      </div>

      <div>
        <h2 class="text-lg text-color">{{ $t('user.shell_setup') }}</h2>
        <pre class="cli-box">{{ usageWithShell }}</pre>
      </div>

      <div>
        <h2 class="text-lg text-color">{{ $t('user.api_usage') }}</h2>
        <pre class="cli-box">{{ usageWithCurl }}</pre>
      </div>

      <div>
        <div class="flex items-center">
          <h2 class="text-lg text-color">{{ $t('user.cli_usage') }}</h2>
          <a :href="cliDownload" target="_blank" class="ml-4 text-link">{{ $t('user.dl_cli') }}</a>
        </div>
        <pre class="cli-box">{{ usageWithCli }}</pre>
      </div>
    </div>
  </Scaffold>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';

export default defineComponent({
  name: 'User',

  components: {
    Button,
    Scaffold,
  },

  setup() {
    const apiClient = useApiClient();
    const token = ref<string | undefined>();

    onMounted(async () => {
      token.value = await apiClient.getToken();
    });

    // eslint-disable-next-line no-restricted-globals
    const address = `${location.protocol}//${location.host}`;

    const usageWithShell = computed(() => {
      let usage = `export WOODPECKER_SERVER="${address}"\n`;
      usage += `export WOODPECKER_TOKEN="${token.value}"\n`;
      return usage;
    });

    const usageWithCurl = `# ${useI18n().t(
      'user.shell_setup_before',
    )}\ncurl -i \${WOODPECKER_SERVER}/api/user -H "Authorization: Bearer \${WOODPECKER_TOKEN}"`;

    const usageWithCli = `# ${useI18n().t('user.shell_setup_before')}\nwoodpecker info`;

    const cliDownload = 'https://github.com/woodpecker-ci/woodpecker/releases';

    return { token, usageWithShell, usageWithCurl, usageWithCli, cliDownload, address };
  },
});
</script>

<style scoped>
.cli-box {
  @apply bg-gray-500 p-2 rounded-md text-white break-words dark:bg-dark-400 dark:text-gray-400;
  white-space: pre-wrap;
}
</style>
