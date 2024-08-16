import { useLocalStorage } from '@vueuse/core';

export function getUserLanguage(): string {
  const browserLocale = navigator.language.split('-')[0];
  const selectedLocale = useLocalStorage('woodpecker:locale', browserLocale).value;

  return selectedLocale;
}

export function truncate(text: string, length: number): string {
  if (text.length <= length) {
    return text;
  }

  return `${text.slice(0, length)}...`;
}
