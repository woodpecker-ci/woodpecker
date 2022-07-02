// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import messages from '@intlify/vite-plugin-vue-i18n/messages';
import { createI18n } from 'vue-i18n';

import { getUserLanguage } from '~/utils/locale';

export const i18n = createI18n({
  locale: getUserLanguage(),
  legacy: false,
  globalInjection: true,
  fallbackLocale: 'en',
  messages,
});
