import ApiClient, { encodeQueryString } from './client';
import {
  Agent,
  Cron,
  OrgPermissions,
  Pipeline,
  PipelineConfig,
  PipelineFeed,
  PipelineLog,
  PipelineStep,
  PullRequest,
  QueueInfo,
  Registry,
  Repo,
  RepoPermissions,
  RepoSettings,
  Secret,
  User,
} from './types';

type RepoListOptions = {
  all?: boolean;
};

// PipelineOptions is the data for creating a new pipeline
type PipelineOptions = {
  branch: string;
  variables: Record<string, string>;
};

type DeploymentOptions = {
  id: string;
  environment: string;
  variables: Record<string, string>;
};

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

  getRepoBranches(repoId: number, page: number): Promise<string[]> {
    return this._get(`/api/repos/${repoId}/branches?page=${page}`) as Promise<string[]>;
  }

  getRepoPullRequests(repoId: number, page: number): Promise<PullRequest[]> {
    return this._get(`/api/repos/${repoId}/pull_requests?page=${page}`) as Promise<PullRequest[]>;
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
  // specified target environment.
  deployPipeline(repoId: number, pipelineNumber: string, options: DeploymentOptions): Promise<Pipeline> {
    const vars = {
      ...options.variables,
      event: 'deployment',
      deploy_to: options.environment,
    };
    const query = encodeQueryString(vars);
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}?${query}`) as Promise<Pipeline>;
  }

  getPipelineList(repoId: number, opts?: Record<string, string | number | boolean>): Promise<Pipeline[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/pipelines?${query}`) as Promise<Pipeline[]>;
  }

  getPipeline(repoId: number, pipelineNumber: number | 'latest'): Promise<Pipeline> {
    return this._get(`/api/repos/${repoId}/pipelines/${pipelineNumber}`) as Promise<Pipeline>;
  }

  getPipelineConfig(repoId: number, pipelineNumber: number): Promise<PipelineConfig[]> {
    return this._get(`/api/repos/${repoId}/pipelines/${pipelineNumber}/config`) as Promise<PipelineConfig[]>;
  }

  getPipelineFeed(opts?: Record<string, string | number | boolean>): Promise<PipelineFeed[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/user/feed?${query}`) as Promise<PipelineFeed[]>;
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

  skipPipelineWorkflow(owner: string, repo: string, pipeline: string, workflowId: number): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/pipelines/${pipeline}/skip/${workflowId}`);
  }

  restartPipeline(
    repoId: number,
    pipeline: string,
    opts?: Record<string, string | number | boolean>,
  ): Promise<Pipeline> {
    const query = encodeQueryString(opts);
    return this._post(`/api/repos/${repoId}/pipelines/${pipeline}?${query}`) as Promise<Pipeline>;
  }

  getLogs(repoId: number, pipeline: number, step: number): Promise<PipelineLog[]> {
    return this._get(`/api/repos/${repoId}/logs/${pipeline}/${step}`) as Promise<PipelineLog[]>;
  }

  getSecretList(repoId: number, page: number): Promise<Secret[] | null> {
    return this._get(`/api/repos/${repoId}/secrets?page=${page}`) as Promise<Secret[] | null>;
  }

  createSecret(repoId: number, secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/secrets`, secret);
  }

  updateSecret(repoId: number, secret: Partial<Secret>): Promise<unknown> {
    return this._patch(`/api/repos/${repoId}/secrets/${secret.name}`, secret);
  }

  deleteSecret(repoId: number, secretName: string): Promise<unknown> {
    return this._delete(`/api/repos/${repoId}/secrets/${secretName}`);
  }

  getRegistryList(repoId: number, page: number): Promise<Registry[] | null> {
    return this._get(`/api/repos/${repoId}/registry?page=${page}`) as Promise<Registry[] | null>;
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

  getCronList(repoId: number, page: number): Promise<Cron[] | null> {
    return this._get(`/api/repos/${repoId}/cron?page=${page}`) as Promise<Cron[] | null>;
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

  getOrgPermissions(owner: string): Promise<OrgPermissions> {
    return this._get(`/api/orgs/${owner}/permissions`) as Promise<OrgPermissions>;
  }

  getOrgSecretList(owner: string, page: number): Promise<Secret[] | null> {
    return this._get(`/api/orgs/${owner}/secrets?page=${page}`) as Promise<Secret[] | null>;
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

  getGlobalSecretList(page: number): Promise<Secret[] | null> {
    return this._get(`/api/secrets?page=${page}`) as Promise<Secret[] | null>;
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

  getAgents(page: number): Promise<Agent[] | null> {
    return this._get(`/api/agents?page=${page}`) as Promise<Agent[] | null>;
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

  getQueueInfo(): Promise<QueueInfo> {
    return this._get('/api/queue/info') as Promise<QueueInfo>;
  }

  pauseQueue(): Promise<unknown> {
    return this._post('/api/queue/pause');
  }

  resumeQueue(): Promise<unknown> {
    return this._post('/api/queue/resume');
  }

  getUsers(page: number): Promise<User[] | null> {
    return this._get(`/api/users?page=${page}`) as Promise<User[] | null>;
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

  // eslint-disable-next-line promise/prefer-await-to-callbacks
  on(callback: (data: { pipeline?: Pipeline; repo?: Repo; step?: PipelineStep }) => void): EventSource {
    return this._subscribe('/stream/events', callback, {
      reconnect: true,
    });
  }

  streamLogs(
    repoId: number,
    pipeline: number,
    step: number,
    // eslint-disable-next-line promise/prefer-await-to-callbacks
    callback: (data: PipelineLog) => void,
  ): EventSource {
    return this._subscribe(`/api/stream/logs/${repoId}/${pipeline}/${step}`, callback, {
      reconnect: true,
    });
  }
}
