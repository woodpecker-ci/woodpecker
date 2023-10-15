import { computed, ref, watch } from 'vue';

export enum Theme {
  Auto = 'auto',
  Light = 'light',
  Dark = 'dark',
}

const LS_THEME = 'woodpecker:theme';
const activeTheme = ref(Theme.Auto);

function resolveAuto(theme: Theme) {
  if (theme === Theme.Auto) {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? Theme.Dark : Theme.Light;
  }
  return theme;
}

watch(activeTheme, (theme) => {
  if (resolveAuto(theme) === Theme.Dark) {
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

function setTheme(theme: Theme) {
  activeTheme.value = theme;
  localStorage.setItem(LS_THEME, theme);
}

function load() {
  const isActive = localStorage.getItem(LS_THEME) as Theme | null;
  if (isActive === null) {
    setTheme(Theme.Auto);
  } else {
    setTheme(isActive);
  }
}

load();

export function useTheme() {
  return {
    darkMode: computed(() => resolveAuto(activeTheme.value) === Theme.Dark),
    theme: computed({
      get() {
        return activeTheme.value;
      },
      set(theme: Theme) {
        setTheme(theme);
      },
    }),
  };
}
