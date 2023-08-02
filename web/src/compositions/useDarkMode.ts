import { computed, ref, watch } from 'vue';

const LS_DARK_MODE = 'woodpecker:dark-mode';
const isDarkModeActive = ref(false);

watch(isDarkModeActive, (isActive) => {
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
});

function setDarkMode(isActive: boolean) {
  isDarkModeActive.value = isActive;
  localStorage.setItem(LS_DARK_MODE, isActive ? 'dark' : 'light');
}

function load() {
  const isActive = localStorage.getItem(LS_DARK_MODE) as 'dark' | 'light' | null;
  if (isActive === null) {
    setDarkMode(window.matchMedia('(prefers-color-scheme: dark)').matches);
  } else {
    setDarkMode(isActive === 'dark');
  }
}

load();

export function useDarkMode() {
  return {
    darkMode: computed({
      get() {
        return isDarkModeActive.value;
      },
      set(isActive: boolean) {
        setDarkMode(isActive);
      },
    }),
  };
}
