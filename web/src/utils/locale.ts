import { useStorage } from '@vueuse/core';

export function getUserLanguage(): string {
  const browserLocale = navigator.language.split('-')[0];
  const selectedLocale = useStorage('woodpecker:locale', browserLocale).value;

  return selectedLocale;
}
