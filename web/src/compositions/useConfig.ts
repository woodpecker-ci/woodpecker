import { User } from '~/lib/api/types';

declare global {
  interface Window {
    WOODPECKER_USER: User | undefined;
    WOODPECKER_SYNC: boolean | undefined;
    WOODPECKER_DOCS: string | undefined;
    WOODPECKER_VERSION: string | undefined;
    WOODPECKER_CSRF: string | undefined;
    WOODPECKER_FORGE: string | undefined;
  }
}

export default () => ({
  user: window.WOODPECKER_USER || null,
  syncing: window.WOODPECKER_SYNC || null,
  docs: window.WOODPECKER_DOCS || null,
  version: window.WOODPECKER_VERSION,
  csrf: window.WOODPECKER_CSRF || null,
  forge: window.WOODPECKER_FORGE || null,
});
