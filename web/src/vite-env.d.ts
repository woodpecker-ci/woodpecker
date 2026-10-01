// / <reference types="vite/client" />

declare module 'virtual:vue-i18n-supported-locales' {
  export const SUPPORTED_LOCALES: string[];
  export const LOCALE_NAMES: Partial<Record<string, string>>;
}
