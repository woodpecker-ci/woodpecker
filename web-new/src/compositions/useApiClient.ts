import WoodpeckerClient from '~/lib/api';
import useConfig from './useConfig';

let apiClient: WoodpeckerClient | undefined;

export default (): WoodpeckerClient => {
  if (!apiClient) {
    const config = useConfig();
    const server = 'http://localhost:8000';
    const token = '';
    const csrf = config.csrf;

    if (!csrf) {
      throw new Error('CSRF unknown');
    }

    apiClient = new WoodpeckerClient(server, token, csrf);
  }

  return apiClient;
};
