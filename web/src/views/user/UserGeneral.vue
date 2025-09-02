<template>
  <Settings :title="$t('user.settings.general.general')">
    <InputField :label="$t('user.settings.general.language')">
      <template #default="{ id }">
        <SelectField :id="id" v-model="selectedLocale" class="mt-2" :options="localeOptions" />
      </template>
      <template #description>
        <i18n-t keypath="help_translating" tag="p">
          <a
            rel="noopener noreferrer"
            href="https://translate.woodpecker-ci.org/projects/woodpecker-ci/ui/"
            target="_blank"
            class="underline"
          >
            {{ $t('weblate') }}
          </a>
        </i18n-t>
      </template>
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
import { useStorage } from '@vueuse/core';
import { SUPPORTED_LOCALES } from 'virtual:vue-i18n-supported-locales';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import Settings from '~/components/layout/Settings.vue';
import { setI18nLanguage } from '~/compositions/useI18n';
import { useTheme } from '~/compositions/useTheme';
import { useWPTitle } from '~/compositions/useWPTitle';

const { locale, t } = useI18n();
const { storeTheme } = useTheme();

const localeOptions = computed(() =>
  SUPPORTED_LOCALES.map((supportedLocale) => ({
    value: supportedLocale,
    text: new Intl.DisplayNames(supportedLocale, { type: 'language' }).of(supportedLocale) || supportedLocale,
  })),
);

const storedLocale = useStorage('woodpecker:locale', locale.value);
const selectedLocale = computed<string>({
  async set(_selectedLocale) {
    await setI18nLanguage(_selectedLocale);
    storedLocale.value = _selectedLocale;
  },
  get() {
    return storedLocale.value;
  },
});

useWPTitle(computed(() => [t('user.settings.general.general'), t('user.settings.settings')]));
</script>
