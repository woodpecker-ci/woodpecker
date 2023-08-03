<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-wp-background-100">
      <h1 class="ml-2 text-xl text-wp-text-100">{{ $t('user.settings.general.general') }}</h1>
    </div>

    <div class="flex flex-col mt-2 mb-4">
      <div class="flex items-center text-wp-text-100 font-bold mb-2">
        <label>{{ $t('user.settings.general.language') }}</label>
      </div>
      <SelectField v-model="selectedLocale" :options="localeOptions" />
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { useLocalStorage } from '@vueuse/core';
import dayjs from 'dayjs';
import TimeAgo from 'javascript-time-ago';
import { SUPPORTED_LOCALES } from 'virtual:vue-i18n-supported-locales';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import SelectField from '~/components/form/SelectField.vue';
import { setI18nLanguage } from '~/compositions/useI18n';

const { locale } = useI18n();

const localeOptions = computed(() =>
  SUPPORTED_LOCALES.map((supportedLocale) => ({
    value: supportedLocale,
    text: new Intl.DisplayNames(supportedLocale, { type: 'language' }).of(supportedLocale) || supportedLocale,
  })),
);

const storedLocale = useLocalStorage('woodpecker:locale', locale.value);
const selectedLocale = computed<string>({
  async set(_selectedLocale) {
    await setI18nLanguage(_selectedLocale);
    storedLocale.value = _selectedLocale;
    dayjs.locale(_selectedLocale);
    TimeAgo.setDefaultLocale(_selectedLocale);
  },
  get() {
    return storedLocale.value;
  },
});
</script>
