<template>
  <FluidContainer class="space-y-4 flex flex-col my-0">
    <Button class="ml-auto" :text="$t('logout')" :to="`${address}/logout`" />

    <div>
      <h2 class="text-lg text-color">{{ $t('user.token') }}</h2>
      <CodeBox>{{ token }}</CodeBox>
    </div>

    <div>
      <h2 class="text-lg text-color">{{ $t('user.shell_setup') }}</h2>
      <CodeBox>{{ usageWithShell }}</CodeBox>
    </div>

    <div>
      <h2 class="text-lg text-color">{{ $t('user.api_usage') }}</h2>
      <CodeBox>{{ usageWithCurl }}</CodeBox>
    </div>

    <div>
      <div class="flex items-center">
        <h2 class="text-lg text-color">{{ $t('user.cli_usage') }}</h2>
        <a :href="cliDownload" target="_blank" class="ml-4 text-link">{{ $t('user.dl_cli') }}</a>
      </div>
      <CodeBox>{{ usageWithCli }}</CodeBox>
    </div>
  </FluidContainer>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import CodeBox from '~/components/layout/CodeBox.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import useApiClient from '~/compositions/useApiClient';

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

const usageWithCurl =
  // eslint-disable-next-line no-template-curly-in-string
  `# ${useI18n().t(
    'user.shell_setup_before',
  )}\ncurl -i \${WOODPECKER_SERVER}/api/user -H "Authorization: Bearer \${WOODPECKER_TOKEN}"`;

const usageWithCli = `# ${useI18n().t('user.shell_setup_before')}\nwoodpecker info`;

const cliDownload = 'https://github.com/woodpecker-ci/woodpecker/releases';
</script>
