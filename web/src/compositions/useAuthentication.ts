import useConfig from '~/compositions/useConfig';

export default () =>
  ({
    isAuthenticated: useConfig().user,

    user: useConfig().user,

    authenticate(origin?: string) {
      const url = `/login?url=${origin || ''}`;
      window.location.href = url;
    },
  } as const);
