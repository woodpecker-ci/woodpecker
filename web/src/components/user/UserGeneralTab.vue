<template>
  <Settings :title="$t('user.settings.general.general')">
    <InputField :label="$t('user.settings.general.language')">
      <SelectField v-model="selectedLocale" :options="localeOptions" />
    </InputField>
    <InputField :label="$t('user.settings.general.theme.theme')">
      <SelectField
        v-model="selectedTheme"
        :options="[
          { value: Theme.Auto, text: $t('user.settings.general.theme.auto') },
          { value: Theme.Light, text: $t('user.settings.general.theme.light') },
          { value: Theme.Dark, text: $t('user.settings.general.theme.dark') },
        ]"
      />
    </InputField>
  </Settings>
</template>

<script lang="ts" setup>
import { useLocalStorage } from '@vueuse/core';
import dayjs from 'dayjs';
import TimeAgo from 'javascript-time-ago';
import { SUPPORTED_LOCALES } from 'virtual:vue-i18n-supported-locales';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import SelectField from '~/components/form/SelectField.vue';
import Settings from '~/components/layout/Settings.vue';
import { setI18nLanguage } from '~/compositions/useI18n';
import { Theme, useTheme } from '~/compositions/useTheme';

const { locale } = useI18n();
const { theme } = useTheme();

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

const selectedTheme = computed<Theme>({
  set(_selectedTheme) {
    theme.value = _selectedTheme;
  },
  get() {
    return theme.value;
  },
});
</script>
