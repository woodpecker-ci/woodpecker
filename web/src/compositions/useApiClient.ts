import WoodpeckerClient from '~/lib/api';

import useConfig from './useConfig';

let apiClient: WoodpeckerClient | undefined;

export default (): WoodpeckerClient => {
  if (!apiClient) {
    const config = useConfig();
    const server = config.rootURL ?? '';
    const token = null;
    const csrf = config.csrf || null;

    apiClient = new WoodpeckerClient(server, token, csrf);
  }

  return apiClient;
};
