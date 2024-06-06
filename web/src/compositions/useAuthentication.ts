import useConfig from '~/compositions/useConfig';
import useUserConfig from '~/compositions/useUserConfig';

export default () =>
  ({
    isAuthenticated: !!useConfig().user,

    user: useConfig().user,

    authenticate(url?: string) {
      if (url !== undefined) {
        const config = useUserConfig();
        config.setUserConfig('redirectUrl', url);
      }
      window.location.href = `${useConfig().rootPath}/login`;
    },
  }) as const;
