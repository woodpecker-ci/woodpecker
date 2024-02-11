import { useLocalStorage } from '@vueuse/core';

export function getUserLanguage(): string {
  const browserLocale = navigator.language.split('-')[0];
  const selectedLocale = useLocalStorage('woodpecker:locale', browserLocale).value;

  return selectedLocale;
}
