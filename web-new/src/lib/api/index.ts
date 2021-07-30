import ApiClient, { encodeQueryString } from './client';
import { Build, BuildProc, BuildLog, Repo } from './types';

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

  activateRepo(owner: string, repo: string) {
    return this._post('/api/repos/' + owner + '/' + repo);
  }

  updateRepo(owner: string, repo: string, data: Record<string, string | number | boolean>) {
    return this._patch('/api/repos/' + owner + '/' + repo, data);
  }

  deleteRepo(owner: string, repo: string) {
    return this._delete('/api/repos/' + owner + '/' + repo);
  }

  getBuildList(owner: string, repo: string, opts?: Record<string, string | number | boolean>): Promise<Build[]> {
    var query = encodeQueryString(opts);
    return this._get('/api/repos/' + owner + '/' + repo + '/builds?' + query);
  }

  getBuild(owner: string, repo: string, number: string | 'latest') {
    return this._get('/api/repos/' + owner + '/' + repo + '/builds/' + number);
  }

  getBuildFeed(opts?: Record<string, string | number | boolean>): Promise<Build[]> {
    var query = encodeQueryString(opts);
    return this._get('/api/user/feed?' + query);
  }

  cancelBuild(owner: string, repo: string, number: number, ppid: number) {
    return this._delete('/api/repos/' + owner + '/' + repo + '/builds/' + number + '/' + ppid);
  }

  approveBuild(owner: string, repo: string, build: string) {
    return this._post('/api/repos/' + owner + '/' + repo + '/builds/' + build + '/approve');
  }

  declineBuild(owner: string, repo: string, build: string) {
    return this._post('/api/repos/' + owner + '/' + repo + '/builds/' + build + '/decline');
  }

  restartBuild(owner: string, repo: string, build: string, opts?: Record<string, string | number | boolean>) {
    var query = encodeQueryString(opts);
    return this._post('/api/repos/' + owner + '/' + repo + '/builds/' + build + '?' + query);
  }

  getLogs(owner: string, repo: string, build: number, proc: number): Promise<BuildLog[]> {
    return this._get('/api/repos/' + owner + '/' + repo + '/logs/' + build + '/' + proc);
  }

  getArtifact(owner: string, repo: string, build: string, proc: string, file: string) {
    return this._get('/api/repos/' + owner + '/' + repo + '/files/' + build + '/' + proc + '/' + file + '?raw=true');
  }

  getArtifactList(owner: string, repo: string, build: string) {
    return this._get('/api/repos/' + owner + '/' + repo + '/files/' + build);
  }

  getSecretList(owner: string, repo: string) {
    return this._get('/api/repos/' + owner + '/' + repo + '/secrets');
  }

  createSecret(owner: string, repo: string, secret: string) {
    return this._post('/api/repos/' + owner + '/' + repo + '/secrets', secret);
  }

  deleteSecret(owner: string, repo: string, secret: string) {
    return this._delete('/api/repos/' + owner + '/' + repo + '/secrets/' + secret);
  }

  getRegistryList(owner: string, repo: string) {
    return this._get('/api/repos/' + owner + '/' + repo + '/registry');
  }

  createRegistry(owner: string, repo: string, registry: string) {
    return this._post('/api/repos/' + owner + '/' + repo + '/registry', registry);
  }

  deleteRegistry(owner: string, repo: string, address: string) {
    return this._delete('/api/repos/' + owner + '/' + repo + '/registry/' + address);
  }

  getSelf() {
    return this._get('/api/user');
  }

  getToken(): Promise<string> {
    return this._post('/api/user/token');
  }

  on(callback: (data: { build?: Build; repo?: Repo; proc?: BuildProc }) => void) {
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
