import { BasicColorSchema, useColorMode } from '@vueuse/core';
import { computed, watch } from 'vue';

const { store: storeTheme, state: resolvedTheme } = useColorMode({
  storageKey: 'woodpecker:theme',
});

watch(storeTheme, updateTheme);

function updateTheme() {
  if (resolvedTheme.value === 'dark') {
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
}

updateTheme();

export function useTheme() {
  return {
    resolvedTheme,
    theme: computed({
      get() {
        return storeTheme.value;
      },
      set(theme: BasicColorSchema) {
        storeTheme.value = theme;
      },
    }),
  };
}
