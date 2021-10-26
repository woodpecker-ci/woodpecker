import ApiClient, { encodeQueryString } from './client';
import { Build, BuildFeed, BuildLog, BuildProc, Registry, Repo, RepoPermissions, RepoSettings, Secret } from './types';

type RepoListOptions = {
  all?: boolean;
  flush?: boolean;
};
export default class WoodpeckerClient extends ApiClient {
  getRepoList(opts?: RepoListOptions): Promise<Repo[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/user/repos?${query}`) as Promise<Repo[]>;
  }

  getRepo(owner: string, repo: string): Promise<Repo> {
    return this._get(`/api/repos/${owner}/${repo}`) as Promise<Repo>;
  }

  getRepoPermissions(owner: string, repo: string): Promise<RepoPermissions> {
    return this._get(`/api/repos/${owner}/${repo}/permissions`) as Promise<RepoPermissions>;
  }

  getRepoBranches(owner: string, repo: string): Promise<string[]> {
    return this._get(`/api/repos/${owner}/${repo}/branches`) as Promise<string[]>;
  }

  activateRepo(owner: string, repo: string): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}`);
  }

  updateRepo(owner: string, repo: string, repoSettings: RepoSettings): Promise<unknown> {
    return this._patch(`/api/repos/${owner}/${repo}`, repoSettings);
  }

  deleteRepo(owner: string, repo: string): Promise<unknown> {
    return this._delete(`/api/repos/${owner}/${repo}`);
  }

  repairRepo(owner: string, repo: string): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/repair`);
  }

  getBuildList(owner: string, repo: string, opts?: Record<string, string | number | boolean>): Promise<Build[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${owner}/${repo}/builds?${query}`) as Promise<Build[]>;
  }

  getBuild(owner: string, repo: string, number: string | 'latest'): Promise<Build> {
    return this._get(`/api/repos/${owner}/${repo}/builds/${number}`) as Promise<Build>;
  }

  getBuildFeed(opts?: Record<string, string | number | boolean>): Promise<BuildFeed[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/user/feed?${query}`) as Promise<BuildFeed[]>;
  }

  cancelBuild(owner: string, repo: string, number: number, ppid: number): Promise<unknown> {
    return this._delete(`/api/repos/${owner}/${repo}/builds/${number}/${ppid}`);
  }

  approveBuild(owner: string, repo: string, build: string): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/builds/${build}/approve`);
  }

  declineBuild(owner: string, repo: string, build: string): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/builds/${build}/decline`);
  }

  restartBuild(
    owner: string,
    repo: string,
    build: string,
    opts?: Record<string, string | number | boolean>,
  ): Promise<unknown> {
    const query = encodeQueryString(opts);
    return this._post(`/api/repos/${owner}/${repo}/builds/${build}?${query}`);
  }

  getLogs(owner: string, repo: string, build: number, proc: number): Promise<BuildLog[]> {
    return this._get(`/api/repos/${owner}/${repo}/logs/${build}/${proc}`) as Promise<BuildLog[]>;
  }

  getArtifact(owner: string, repo: string, build: string, proc: string, file: string): Promise<unknown> {
    return this._get(`/api/repos/${owner}/${repo}/files/${build}/${proc}/${file}?raw=true`);
  }

  getArtifactList(owner: string, repo: string, build: string): Promise<unknown> {
    return this._get(`/api/repos/${owner}/${repo}/files/${build}`);
  }

  getSecretList(owner: string, repo: string): Promise<Secret[]> {
    return this._get(`/api/repos/${owner}/${repo}/secrets`) as Promise<Secret[]>;
  }

  createSecret(owner: string, repo: string, secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/secrets`, secret);
  }

  deleteSecret(owner: string, repo: string, secretName: string): Promise<unknown> {
    return this._delete(`/api/repos/${owner}/${repo}/secrets/${secretName}`);
  }

  getRegistryList(owner: string, repo: string): Promise<Registry[]> {
    return this._get(`/api/repos/${owner}/${repo}/registry`) as Promise<Registry[]>;
  }

  createRegistry(owner: string, repo: string, registry: Partial<Registry>): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/registry`, registry);
  }

  deleteRegistry(owner: string, repo: string, registryAddress: string): Promise<unknown> {
    return this._delete(`/api/repos/${owner}/${repo}/registry/${registryAddress}`);
  }

  getSelf(): Promise<unknown> {
    return this._get('/api/user');
  }

  getToken(): Promise<string> {
    return this._post('/api/user/token') as Promise<string>;
  }

  // eslint-disable-next-line promise/prefer-await-to-callbacks
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
    // eslint-disable-next-line promise/prefer-await-to-callbacks
    callback: (data: BuildLog) => void,
  ): EventSource {
    return this._subscribe(`/stream/logs/${owner}/${repo}/${build}/${proc}`, callback, {
      reconnect: true,
    });
  }
}
