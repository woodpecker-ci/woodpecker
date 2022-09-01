import ApiClient, { encodeQueryString } from './client';
import {
  Build,
  BuildConfig,
  BuildFeed,
  BuildLog,
  BuildProc,
  OrgPermissions,
  Registry,
  Repo,
  RepoPermissions,
  RepoSettings,
  Secret,
} from './types';
import { Cron } from './types/cron';

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

  deleteRepo(owner: string, repo: string, remove = true): Promise<unknown> {
    const query = encodeQueryString({ remove });
    return this._delete(`/api/repos/${owner}/${repo}?${query}`);
  }

  repairRepo(owner: string, repo: string): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/repair`);
  }

  getBuildList(owner: string, repo: string, opts?: Record<string, string | number | boolean>): Promise<Build[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${owner}/${repo}/builds?${query}`) as Promise<Build[]>;
  }

  getBuild(owner: string, repo: string, number: number | 'latest'): Promise<Build> {
    return this._get(`/api/repos/${owner}/${repo}/builds/${number}`) as Promise<Build>;
  }

  getBuildConfig(owner: string, repo: string, number: number): Promise<BuildConfig[]> {
    return this._get(`/api/repos/${owner}/${repo}/builds/${number}/config`) as Promise<BuildConfig[]>;
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

  updateSecret(owner: string, repo: string, secret: Partial<Secret>): Promise<unknown> {
    return this._patch(`/api/repos/${owner}/${repo}/secrets/${secret.name}`, secret);
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

  updateRegistry(owner: string, repo: string, registry: Partial<Registry>): Promise<unknown> {
    return this._patch(`/api/repos/${owner}/${repo}/registry/${registry.address}`, registry);
  }

  deleteRegistry(owner: string, repo: string, registryAddress: string): Promise<unknown> {
    return this._delete(`/api/repos/${owner}/${repo}/registry/${registryAddress}`);
  }

  getCronList(owner: string, repo: string): Promise<Cron[]> {
    return this._get(`/api/repos/${owner}/${repo}/cron`) as Promise<Cron[]>;
  }

  createCron(owner: string, repo: string, cron: Partial<Cron>): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/cron`, cron);
  }

  updateCron(owner: string, repo: string, cron: Partial<Cron>): Promise<unknown> {
    return this._patch(`/api/repos/${owner}/${repo}/cron/${cron.id}`, cron);
  }

  deleteCron(owner: string, repo: string, cronId: number): Promise<unknown> {
    return this._delete(`/api/repos/${owner}/${repo}/cron/${cronId}`);
  }

  getOrgPermissions(owner: string): Promise<OrgPermissions> {
    return this._get(`/api/orgs/${owner}/permissions`) as Promise<OrgPermissions>;
  }

  getOrgSecretList(owner: string): Promise<Secret[]> {
    return this._get(`/api/orgs/${owner}/secrets`) as Promise<Secret[]>;
  }

  createOrgSecret(owner: string, secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/orgs/${owner}/secrets`, secret);
  }

  updateOrgSecret(owner: string, secret: Partial<Secret>): Promise<unknown> {
    return this._patch(`/api/orgs/${owner}/secrets/${secret.name}`, secret);
  }

  deleteOrgSecret(owner: string, secretName: string): Promise<unknown> {
    return this._delete(`/api/orgs/${owner}/secrets/${secretName}`);
  }

  getGlobalSecretList(): Promise<Secret[]> {
    return this._get(`/api/secrets`) as Promise<Secret[]>;
  }

  createGlobalSecret(secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/secrets`, secret);
  }

  updateGlobalSecret(secret: Partial<Secret>): Promise<unknown> {
    return this._patch(`/api/secrets/${secret.name}`, secret);
  }

  deleteGlobalSecret(secretName: string): Promise<unknown> {
    return this._delete(`/api/secrets/${secretName}`);
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
