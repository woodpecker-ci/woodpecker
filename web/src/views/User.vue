<template>
  <FluidContainer class="space-y-4 flex flex-col my-0">
    <Button class="ml-auto" :text="$t('logout')" :to="`${address}/logout`" />

    <SelectField v-model="selectedLocale" :options="localeOptions" />

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
  </FluidContainer>
</template>

<script lang="ts" setup>
import { useLocalStorage } from '@vueuse/core';
import dayjs from 'dayjs';
import TimeAgo from 'javascript-time-ago';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import SelectField from '~/components/form/SelectField.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import useApiClient from '~/compositions/useApiClient';

const { t, availableLocales, locale } = useI18n();

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

const usageWithCurl = `# ${t(
  'user.shell_setup_before',
)}\ncurl -i \${WOODPECKER_SERVER}/api/user -H "Authorization: Bearer \${WOODPECKER_TOKEN}"`;

const usageWithCli = `# ${t('user.shell_setup_before')}\nwoodpecker info`;

const cliDownload = 'https://github.com/woodpecker-ci/woodpecker/releases';

const localeOptions = computed(() =>
  availableLocales.map((availableLocale) => ({
    value: availableLocale,
    text: new Intl.DisplayNames(availableLocale, { type: 'language' }).of(availableLocale) || availableLocale,
  })),
);

const storedLocale = useLocalStorage('woodpecker:locale', locale.value);
const selectedLocale = computed<string>({
  set(_selectedLocale) {
    storedLocale.value = _selectedLocale;
    locale.value = _selectedLocale;
    dayjs.locale(_selectedLocale);
    TimeAgo.setDefaultLocale(_selectedLocale);
  },
  get() {
    return storedLocale.value;
  },
});
</script>

<style scoped>
.cli-box {
  @apply bg-gray-500 p-2 rounded-md text-white break-words dark:bg-dark-400 dark:text-gray-400;
  white-space: pre-wrap;
}
</style>
