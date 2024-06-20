import useConfig from '~/compositions/useConfig';
import useUserConfig from '~/compositions/useUserConfig';

export default () =>
  ({
    isAuthenticated: !!useConfig().user,

    user: useConfig().user,

    authenticate(url?: string, forgeId?: number) {
      if (url !== undefined) {
        const config = useUserConfig();
        config.setUserConfig('redirectUrl', url);
      }
      window.location.href = `${useConfig().rootPath}/authorize?forge_id=${forgeId}`;
    },
  }) as const;
