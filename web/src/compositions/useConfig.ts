import { User } from '~/lib/api/types';

declare global {
  interface Window {
    WOODPECKER_USER: User | undefined;
    WOODPECKER_DOCS: string | undefined;
    WOODPECKER_VERSION: string | undefined;
    WOODPECKER_CSRF: string | undefined;
    WOODPECKER_FORGE: string | undefined;
    WOODPECKER_ROOT_PATH: string | undefined;
  }
}

export default () => ({
  user: window.WOODPECKER_USER || null,
  docs: window.WOODPECKER_DOCS || null,
  version: window.WOODPECKER_VERSION,
  csrf: window.WOODPECKER_CSRF || null,
  forge: window.WOODPECKER_FORGE || null,
  rootPath: window.WOODPECKER_ROOT_PATH || '',
});
