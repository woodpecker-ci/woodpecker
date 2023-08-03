import { User } from '~/lib/api/types';

declare global {
  interface Window {
    WOODPECKER_USER: User | undefined;
    WOODPECKER_DOCS: string | undefined;
    WOODPECKER_VERSION: string | undefined;
    WOODPECKER_CSRF: string | undefined;
    WOODPECKER_FORGE: string | undefined;
    WOODPECKER_ROOT_URL: string | undefined;
    WOODPECKER_ENABLE_SWAGGER: boolean | undefined;
  }
}

export default () => ({
  user: window.WOODPECKER_USER || null,
  docs: window.WOODPECKER_DOCS || null,
  version: window.WOODPECKER_VERSION,
  csrf: window.WOODPECKER_CSRF || null,
  forge: window.WOODPECKER_FORGE || null,
  rootURL: window.WOODPECKER_ROOT_URL || null,
  enableSwagger: window.WOODPECKER_ENABLE_SWAGGER || false,
});
