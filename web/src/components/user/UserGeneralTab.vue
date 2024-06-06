<template>
  <Settings :title="$t('user.settings.general.general')">
    <InputField v-slot="{ id }" :label="$t('user.settings.general.language')">
      <SelectField :id="id" v-model="selectedLocale" :options="localeOptions" />
    </InputField>
    <InputField v-slot="{ id }" :label="$t('user.settings.general.theme.theme')">
      <SelectField
        :id="id"
        v-model="storeTheme"
        :options="[
          { value: 'auto', text: $t('user.settings.general.theme.auto') },
          { value: 'light', text: $t('user.settings.general.theme.light') },
          { value: 'dark', text: $t('user.settings.general.theme.dark') },
        ]"
      />
    </InputField>
  </Settings>
</template>

<script lang="ts" setup>
import { useLocalStorage } from '@vueuse/core';
import { SUPPORTED_LOCALES } from 'virtual:vue-i18n-supported-locales';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import SelectField from '~/components/form/SelectField.vue';
import Settings from '~/components/layout/Settings.vue';
import { setI18nLanguage } from '~/compositions/useI18n';
import { useTheme } from '~/compositions/useTheme';

const { locale } = useI18n();
const { storeTheme } = useTheme();

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
  },
  get() {
    return storedLocale.value;
  },
});
</script>
