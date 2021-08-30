import ApiClient, { encodeQueryString } from './client';
import { Build, BuildProc, BuildLog, Repo, BuildFeed, Secret, Registry, RepoSettings } from './types';

type RepoListOptions = {
  all?: boolean;
  flush?: boolean;
};
export default class WoodpeckerClient extends ApiClient {
  constructor(server: string, token: string | null, csrf: string | null) {
    super(server, token, csrf);
  }

  getRepoList(opts?: RepoListOptions): Promise<Repo[]> {
    var query = encodeQueryString(opts);
    return this._get('/api/user/repos?' + query);
  }

  getRepo(owner: string, repo: string): Promise<Repo> {
    return this._get('/api/repos/' + owner + '/' + repo);
  }

  activateRepo(owner: string, repo: string): Promise<unknown> {
    return this._post('/api/repos/' + owner + '/' + repo);
  }

  updateRepo(owner: string, repo: string, repoSettings: RepoSettings): Promise<unknown> {
    return this._patch('/api/repos/' + owner + '/' + repo, repoSettings);
  }

  deleteRepo(owner: string, repo: string): Promise<unknown> {
    return this._delete('/api/repos/' + owner + '/' + repo);
  }

  getBuildList(owner: string, repo: string, opts?: Record<string, string | number | boolean>): Promise<Build[]> {
    var query = encodeQueryString(opts);
    return this._get('/api/repos/' + owner + '/' + repo + '/builds?' + query);
  }

  getBuild(owner: string, repo: string, number: string | 'latest'): Promise<Build> {
    return this._get('/api/repos/' + owner + '/' + repo + '/builds/' + number);
  }

  getBuildFeed(opts?: Record<string, string | number | boolean>): Promise<BuildFeed[]> {
    var query = encodeQueryString(opts);
    return this._get('/api/user/feed?' + query);
  }

  cancelBuild(owner: string, repo: string, number: number, ppid: number): Promise<unknown> {
    return this._delete('/api/repos/' + owner + '/' + repo + '/builds/' + number + '/' + ppid);
  }

  approveBuild(owner: string, repo: string, build: string): Promise<unknown> {
    return this._post('/api/repos/' + owner + '/' + repo + '/builds/' + build + '/approve');
  }

  declineBuild(owner: string, repo: string, build: string): Promise<unknown> {
    return this._post('/api/repos/' + owner + '/' + repo + '/builds/' + build + '/decline');
  }

  restartBuild(
    owner: string,
    repo: string,
    build: string,
    opts?: Record<string, string | number | boolean>,
  ): Promise<unknown> {
    var query = encodeQueryString(opts);
    return this._post('/api/repos/' + owner + '/' + repo + '/builds/' + build + '?' + query);
  }

  getLogs(owner: string, repo: string, build: number, proc: number): Promise<BuildLog[]> {
    return this._get('/api/repos/' + owner + '/' + repo + '/logs/' + build + '/' + proc);
  }

  getArtifact(owner: string, repo: string, build: string, proc: string, file: string): Promise<unknown> {
    return this._get('/api/repos/' + owner + '/' + repo + '/files/' + build + '/' + proc + '/' + file + '?raw=true');
  }

  getArtifactList(owner: string, repo: string, build: string): Promise<unknown> {
    return this._get('/api/repos/' + owner + '/' + repo + '/files/' + build);
  }

  getSecretList(owner: string, repo: string): Promise<Secret[]> {
    return this._get('/api/repos/' + owner + '/' + repo + '/secrets');
  }

  createSecret(owner: string, repo: string, secret: Partial<Secret>): Promise<unknown> {
    return this._post('/api/repos/' + owner + '/' + repo + '/secrets', secret);
  }

  deleteSecret(owner: string, repo: string, secretName: string): Promise<unknown> {
    return this._delete('/api/repos/' + owner + '/' + repo + '/secrets/' + secretName);
  }

  getRegistryList(owner: string, repo: string): Promise<Registry[]> {
    return this._get('/api/repos/' + owner + '/' + repo + '/registry');
  }

  createRegistry(owner: string, repo: string, registry: Partial<Registry>): Promise<unknown> {
    return this._post('/api/repos/' + owner + '/' + repo + '/registry', registry);
  }

  deleteRegistry(owner: string, repo: string, registryAddress: string): Promise<unknown> {
    return this._delete('/api/repos/' + owner + '/' + repo + '/registry/' + registryAddress);
  }

  getSelf(): Promise<unknown> {
    return this._get('/api/user');
  }

  getToken(): Promise<string> {
    return this._post('/api/user/token');
  }

  on(callback: (data: { build?: Build; repo?: Repo; proc?: BuildProc }) => void): EventSource {
    return this._subscribe('/stream/events', callback, {
      reconnect: true,
    });
  }

  streamLogs(
    owner: string,
    repo: string,
    build: number,
    proc: number,
    callback: (data: BuildLog) => void,
  ): EventSource {
    return this._subscribe('/stream/logs/' + owner + '/' + repo + '/' + build + '/' + proc, callback, {
      reconnect: true,
    });
  }
}
