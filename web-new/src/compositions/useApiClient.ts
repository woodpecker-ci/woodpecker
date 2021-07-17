import WoodpeckerClient from '~/lib/api';

let apiClient: WoodpeckerClient | undefined;

export default (): WoodpeckerClient => {
  if (!apiClient) {
    const server = 'http://localhost:8000';
    const token = '';
    const csrf = '';

    apiClient = new WoodpeckerClient(server, token, csrf);
  }

  return apiClient;
};
