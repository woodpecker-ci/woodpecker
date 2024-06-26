import ApiClient, { encodeQueryString } from './client';
import type {
  Agent,
  Cron,
  Forge,
  Org,
  OrgPermissions,
  Pipeline,
  PipelineConfig,
  PipelineFeed,
  PipelineLog,
  PullRequest,
  QueueInfo,
  Registry,
  Repo,
  RepoPermissions,
  RepoSettings,
  Secret,
  User,
} from './types';

interface RepoListOptions {
  all?: boolean;
}

// PipelineOptions is the data for creating a new pipeline
interface PipelineOptions {
  branch: string;
  variables: Record<string, string>;
}

interface DeploymentOptions {
  id: string;
  environment: string;
  task: string;
  variables: Record<string, string>;
}

interface PaginationOptions {
  page?: number;
  perPage?: number;
}

export default class WoodpeckerClient extends ApiClient {
  getRepoList(opts?: RepoListOptions): Promise<Repo[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/user/repos?${query}`) as Promise<Repo[]>;
  }

  lookupRepo(owner: string, name: string): Promise<Repo | undefined> {
    return this._get(`/api/repos/lookup/${owner}/${name}`) as Promise<Repo | undefined>;
  }

  getRepo(repoId: number): Promise<Repo> {
    return this._get(`/api/repos/${repoId}`) as Promise<Repo>;
  }

  getRepoPermissions(repoId: number): Promise<RepoPermissions> {
    return this._get(`/api/repos/${repoId}/permissions`) as Promise<RepoPermissions>;
  }

  getRepoBranches(repoId: number, opts?: PaginationOptions): Promise<string[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/branches?${query}`) as Promise<string[]>;
  }

  getRepoPullRequests(repoId: number, opts?: PaginationOptions): Promise<PullRequest[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/pull_requests?${query}`) as Promise<PullRequest[]>;
  }

  activateRepo(forgeRemoteId: string): Promise<Repo> {
    return this._post(`/api/repos?forge_remote_id=${forgeRemoteId}`) as Promise<Repo>;
  }

  updateRepo(repoId: number, repoSettings: RepoSettings): Promise<unknown> {
    return this._patch(`/api/repos/${repoId}`, repoSettings);
  }

  deleteRepo(repoId: number, remove = true): Promise<unknown> {
    const query = encodeQueryString({ remove });
    return this._delete(`/api/repos/${repoId}?${query}`);
  }

  repairRepo(repoId: number): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/repair`);
  }

  createPipeline(repoId: number, options: PipelineOptions): Promise<Pipeline> {
    return this._post(`/api/repos/${repoId}/pipelines`, options) as Promise<Pipeline>;
  }

  // Deploy triggers a deployment for an existing pipeline using the
  // specified target environment and task.
  deployPipeline(repoId: number, pipelineNumber: string, options: DeploymentOptions): Promise<Pipeline> {
    const vars = {
      ...options.variables,
      event: 'deployment',
      deploy_to: options.environment,
      deploy_task: options.task,
    };
    const query = encodeQueryString(vars);
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}?${query}`) as Promise<Pipeline>;
  }

  getPipelineList(repoId: number, opts?: PaginationOptions & { before?: string; after?: string }): Promise<Pipeline[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/pipelines?${query}`) as Promise<Pipeline[]>;
  }

  getPipeline(repoId: number, pipelineNumber: number | 'latest'): Promise<Pipeline> {
    return this._get(`/api/repos/${repoId}/pipelines/${pipelineNumber}`) as Promise<Pipeline>;
  }

  getPipelineConfig(repoId: number, pipelineNumber: number): Promise<PipelineConfig[]> {
    return this._get(`/api/repos/${repoId}/pipelines/${pipelineNumber}/config`) as Promise<PipelineConfig[]>;
  }

  getPipelineFeed(): Promise<PipelineFeed[]> {
    return this._get(`/api/user/feed`) as Promise<PipelineFeed[]>;
  }

  cancelPipeline(repoId: number, pipelineNumber: number): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}/cancel`);
  }

  approvePipeline(repoId: number, pipelineNumber: string): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}/approve`);
  }

  declinePipeline(repoId: number, pipelineNumber: string): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}/decline`);
  }

  restartPipeline(
    repoId: number,
    pipeline: string,
    opts?: { event?: string; deploy_to?: string; fork?: boolean },
  ): Promise<Pipeline> {
    const query = encodeQueryString(opts);
    return this._post(`/api/repos/${repoId}/pipelines/${pipeline}?${query}`) as Promise<Pipeline>;
  }

  getLogs(repoId: number, pipeline: number, step: number): Promise<PipelineLog[]> {
    return this._get(`/api/repos/${repoId}/logs/${pipeline}/${step}`) as Promise<PipelineLog[]>;
  }

  deleteLogs(repoId: number, pipeline: number, step: number): Promise<unknown> {
    return this._delete(`/api/repos/${repoId}/logs/${pipeline}/${step}`);
  }

  getSecretList(repoId: number, opts?: PaginationOptions): Promise<Secret[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/secrets?${query}`) as Promise<Secret[] | null>;
  }

  createSecret(repoId: number, secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/secrets`, secret);
  }

  updateSecret(repoId: number, secret: Partial<Secret>): Promise<unknown> {
    const secretName = encodeURIComponent(secret.name ?? '');
    return this._patch(`/api/repos/${repoId}/secrets/${secretName}`, secret);
  }

  deleteSecret(repoId: number, secretName: string): Promise<unknown> {
    const name = encodeURIComponent(secretName);
    return this._delete(`/api/repos/${repoId}/secrets/${name}`);
  }

  getRegistryList(repoId: number, opts?: PaginationOptions): Promise<Registry[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/registry?${query}`) as Promise<Registry[] | null>;
  }

  createRegistry(repoId: number, registry: Partial<Registry>): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/registry`, registry);
  }

  updateRegistry(repoId: number, registry: Partial<Registry>): Promise<unknown> {
    return this._patch(`/api/repos/${repoId}/registry/${registry.address}`, registry);
  }

  deleteRegistry(repoId: number, registryAddress: string): Promise<unknown> {
    return this._delete(`/api/repos/${repoId}/registry/${registryAddress}`);
  }

  getCronList(repoId: number, opts?: PaginationOptions): Promise<Cron[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/cron?${query}`) as Promise<Cron[] | null>;
  }

  createCron(repoId: number, cron: Partial<Cron>): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/cron`, cron);
  }

  updateCron(repoId: number, cron: Partial<Cron>): Promise<unknown> {
    return this._patch(`/api/repos/${repoId}/cron/${cron.id}`, cron);
  }

  deleteCron(repoId: number, cronId: number): Promise<unknown> {
    return this._delete(`/api/repos/${repoId}/cron/${cronId}`);
  }

  runCron(repoId: number, cronId: number): Promise<Pipeline> {
    return this._post(`/api/repos/${repoId}/cron/${cronId}`) as Promise<Pipeline>;
  }

  getOrg(orgId: number): Promise<Org> {
    return this._get(`/api/orgs/${orgId}`) as Promise<Org>;
  }

  lookupOrg(name: string): Promise<Org> {
    return this._get(`/api/orgs/lookup/${name}`) as Promise<Org>;
  }

  getOrgPermissions(orgId: number): Promise<OrgPermissions> {
    return this._get(`/api/orgs/${orgId}/permissions`) as Promise<OrgPermissions>;
  }

  getOrgSecretList(orgId: number, opts?: PaginationOptions): Promise<Secret[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/orgs/${orgId}/secrets?${query}`) as Promise<Secret[] | null>;
  }

  createOrgSecret(orgId: number, secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/orgs/${orgId}/secrets`, secret);
  }

  updateOrgSecret(orgId: number, secret: Partial<Secret>): Promise<unknown> {
    const secretName = encodeURIComponent(secret.name ?? '');
    return this._patch(`/api/orgs/${orgId}/secrets/${secretName}`, secret);
  }

  deleteOrgSecret(orgId: number, secretName: string): Promise<unknown> {
    const name = encodeURIComponent(secretName);
    return this._delete(`/api/orgs/${orgId}/secrets/${name}`);
  }

  getGlobalSecretList(opts?: PaginationOptions): Promise<Secret[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/secrets?${query}`) as Promise<Secret[] | null>;
  }

  createGlobalSecret(secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/secrets`, secret);
  }

  updateGlobalSecret(secret: Partial<Secret>): Promise<unknown> {
    const secretName = encodeURIComponent(secret.name ?? '');
    return this._patch(`/api/secrets/${secretName}`, secret);
  }

  deleteGlobalSecret(secretName: string): Promise<unknown> {
    const name = encodeURIComponent(secretName);
    return this._delete(`/api/secrets/${name}`);
  }

  getSelf(): Promise<unknown> {
    return this._get('/api/user');
  }

  getToken(): Promise<string> {
    return this._post('/api/user/token') as Promise<string>;
  }

  getAgents(opts?: PaginationOptions): Promise<Agent[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/agents?${query}`) as Promise<Agent[] | null>;
  }

  getAgent(agentId: Agent['id']): Promise<Agent> {
    return this._get(`/api/agents/${agentId}`) as Promise<Agent>;
  }

  createAgent(agent: Partial<Agent>): Promise<Agent> {
    return this._post('/api/agents', agent) as Promise<Agent>;
  }

  updateAgent(agent: Partial<Agent>): Promise<unknown> {
    return this._patch(`/api/agents/${agent.id}`, agent);
  }

  deleteAgent(agent: Agent): Promise<unknown> {
    return this._delete(`/api/agents/${agent.id}`);
  }

  getForges(opts?: PaginationOptions): Promise<Forge[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/forges?${query}`) as Promise<Forge[] | null>;
  }

  getForge(forgeId: Forge['id']): Promise<Forge> {
    return this._get(`/api/forges/${forgeId}`) as Promise<Forge>;
  }

  createForge(forge: Partial<Forge>): Promise<Forge> {
    return this._post('/api/forges', forge) as Promise<Forge>;
  }

  updateForge(forge: Partial<Forge>): Promise<unknown> {
    return this._patch(`/api/forges/${forge.id}`, forge);
  }

  deleteForge(forge: Forge): Promise<unknown> {
    return this._delete(`/api/forges/${forge.id}`);
  }

  getQueueInfo(): Promise<QueueInfo> {
    return this._get('/api/queue/info') as Promise<QueueInfo>;
  }

  pauseQueue(): Promise<unknown> {
    return this._post('/api/queue/pause');
  }

  resumeQueue(): Promise<unknown> {
    return this._post('/api/queue/resume');
  }

  getUsers(opts?: PaginationOptions): Promise<User[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/users?${query}`) as Promise<User[] | null>;
  }

  getUser(username: string): Promise<User> {
    return this._get(`/api/users/${username}`) as Promise<User>;
  }

  createUser(user: Partial<User>): Promise<User> {
    return this._post('/api/users', user) as Promise<User>;
  }

  updateUser(user: Partial<User>): Promise<unknown> {
    return this._patch(`/api/users/${user.login}`, user);
  }

  deleteUser(user: User): Promise<unknown> {
    return this._delete(`/api/users/${user.login}`);
  }

  resetToken(): Promise<string> {
    return this._delete('/api/user/token') as Promise<string>;
  }

  getOrgs(opts?: PaginationOptions): Promise<Org[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/orgs?${query}`) as Promise<Org[] | null>;
  }

  deleteOrg(org: Org): Promise<unknown> {
    return this._delete(`/api/orgs/${org.id}`);
  }

  getAllRepos(opts?: PaginationOptions): Promise<Repo[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos?${query}`) as Promise<Repo[] | null>;
  }

  repairAllRepos(): Promise<unknown> {
    return this._post(`/api/repos/repair`);
  }

  // TODO enable again with eslint-plugin-promise eslint-disable-next-line promise/prefer-await-to-callbacks
  on(callback: (data: { pipeline?: Pipeline; repo?: Repo }) => void): EventSource {
    return this._subscribe('/api/stream/events', callback, {
      reconnect: true,
    });
  }

  streamLogs(
    repoId: number,
    pipeline: number,
    step: number,
    // TODO enable again with eslint-plugin-promise eslint-disable-next-line promise/prefer-await-to-callbacks
    callback: (data: PipelineLog) => void,
  ): EventSource {
    return this._subscribe(`/api/stream/logs/${repoId}/${pipeline}/${step}`, callback, {
      reconnect: true,
    });
  }
}
