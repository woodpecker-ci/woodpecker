import { User } from '~/lib/api/types';

declare global {
  interface Window {
    WOODPECKER_USER: User | undefined;
    WOODPECKER_VERSION: string | undefined;
    WOODPECKER_CSRF: string | undefined;
    WOODPECKER_FORGE: 'github' | 'gitlab' | 'gitea' | 'bitbucket' | undefined;
    WOODPECKER_ROOT_PATH: string | undefined;
    WOODPECKER_ENABLE_SWAGGER: boolean | undefined;
  }
}

export default () => ({
  user: window.WOODPECKER_USER || null,
  version: window.WOODPECKER_VERSION,
  csrf: window.WOODPECKER_CSRF || null,
  forge: window.WOODPECKER_FORGE || null,
  rootPath: window.WOODPECKER_ROOT_PATH || '',
  enableSwagger: window.WOODPECKER_ENABLE_SWAGGER || false,
});
