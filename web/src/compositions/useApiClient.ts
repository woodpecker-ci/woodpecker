import useConfig from './useConfig';
import WoodpeckerClient from '~/lib/api';


let apiClient: WoodpeckerClient | undefined;

export default (): WoodpeckerClient => {
  if (!apiClient) {
    const config = useConfig();
    const server = config.rootPath;
    const token = null;
    const csrf = config.csrf || null;

    apiClient = new WoodpeckerClient(server, token, csrf);
  }

  return apiClient;
};
