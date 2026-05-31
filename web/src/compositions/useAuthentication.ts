import useConfig from '~/compositions/useConfig';

import { closeEvents } from './useEvents';

export default () =>
  ({
    isAuthenticated: !!useConfig().user,

    user: useConfig().user,

    authenticate(forgeId?: number) {
      closeEvents();
      window.location.href = `${useConfig().rootPath}/authorize?${forgeId !== undefined ? `forge_id=${forgeId}` : ''}`;
    },
  }) as const;
