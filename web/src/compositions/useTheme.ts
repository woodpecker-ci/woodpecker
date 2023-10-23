import { BasicColorSchema, useColorMode } from '@vueuse/core';
import { computed, watch } from 'vue';

const { system, store } = useColorMode();
const resolvedTheme = computed(() => (store.value === 'auto' ? system.value : store.value));

watch(store, () => {
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
});

function setTheme(theme: BasicColorSchema) {
  store.value = theme;
}

setTheme(store.value);

export function useTheme() {
  return {
    darkMode: computed(() => resolvedTheme.value === 'dark'),
    theme: computed({
      get() {
        return store.value;
      },
      set(theme: BasicColorSchema) {
        setTheme(theme);
      },
    }),
  };
}
