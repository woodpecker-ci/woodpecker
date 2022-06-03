// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore
import messages from '@intlify/vite-plugin-vue-i18n/messages';
import { createI18n } from 'vue-i18n';

export const i18n = createI18n({
  locale: navigator.language.split('-')[0],
  legacy: false,
  globalInjection: true,
  fallbackLocale: 'en',
  messages,
});
