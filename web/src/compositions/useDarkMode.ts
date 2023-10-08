import { useStorage } from '@vueuse/core';
import { watch } from 'vue';

const isDarkModeActive = useStorage('woodpecker:dark-mode', window.matchMedia('(prefers-color-scheme: dark)').matches);

watch(
  isDarkModeActive,
  (isActive) => {
    if (isActive) {
      document.documentElement.classList.remove('light');
      document.documentElement.classList.add('dark');
      document.documentElement.setAttribute('data-theme', 'dark');
      document.querySelector('meta[name=theme-color]')?.setAttribute('content', '#2A2E3A'); // internal-wp-secondary-600 (see windi.config.ts)
    } else {
      document.documentElement.classList.remove('dark');
      document.documentElement.classList.add('light');
      document.documentElement.setAttribute('data-theme', 'light');
      document.querySelector('meta[name=theme-color]')?.setAttribute('content', '#369943'); // internal-wp-primary-400
    }
  },
  { immediate: true },
);

export function useDarkMode() {
  return {
    darkMode: isDarkModeActive,
  };
}
