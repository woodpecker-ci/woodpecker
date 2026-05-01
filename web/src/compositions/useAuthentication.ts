import useConfig from '~/compositions/useConfig';

export default () =>
  ({
    isAuthenticated: !!useConfig().user,

    user: useConfig().user,

    authenticate(forgeId?: number) {
      window.location.href = `${useConfig().rootPath}/authorize?${forgeId !== undefined ? `forge_id=${forgeId}` : ''}`;
    },
  }) as const;
