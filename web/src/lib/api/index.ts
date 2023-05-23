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

  getRepo(owner: string, repo: string): Promise<Repo> {
    return this._get(`/api/repos/${owner}/${repo}`) as Promise<Repo>;
  }

  getRepoPermissions(owner: string, repo: string): Promise<RepoPermissions> {
    return this._get(`/api/repos/${owner}/${repo}/permissions`) as Promise<RepoPermissions>;
  }

  getRepoBranches(owner: string, repo: string, page: number): Promise<string[]> {
    return this._get(`/api/repos/${owner}/${repo}/branches?page=${page}`) as Promise<string[]>;
  }

  getRepoPullRequests(owner: string, repo: string, page: number): Promise<PullRequest[]> {
    return this._get(`/api/repos/${owner}/${repo}/pull_requests?page=${page}`) as Promise<PullRequest[]>;
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

  createPipeline(owner: string, repo: string, options: PipelineOptions): Promise<Pipeline> {
    return this._post(`/api/repos/${owner}/${repo}/pipelines`, options) as Promise<Pipeline>;
  }

  // Deploy triggers a deployment for an existing pipeline using the
  // specified target environment.
  deployPipeline(owner: string, repo: string, pipelineNumber: string, options: DeploymentOptions): Promise<Pipeline> {
    const vars = {
      ...options.variables,
      event: 'deployment',
      deploy_to: options.environment,
    };
    const query = encodeQueryString(vars);
    return this._post(`/api/repos/${owner}/${repo}/pipelines/${pipelineNumber}?${query}`) as Promise<Pipeline>;
  }

  getPipelineList(owner: string, repo: string, opts?: Record<string, string | number | boolean>): Promise<Pipeline[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${owner}/${repo}/pipelines?${query}`) as Promise<Pipeline[]>;
  }

  getPipeline(owner: string, repo: string, pipelineNumber: number | 'latest'): Promise<Pipeline> {
    return this._get(`/api/repos/${owner}/${repo}/pipelines/${pipelineNumber}`) as Promise<Pipeline>;
  }

  getPipelineConfig(owner: string, repo: string, pipelineNumber: number): Promise<PipelineConfig[]> {
    return this._get(`/api/repos/${owner}/${repo}/pipelines/${pipelineNumber}/config`) as Promise<PipelineConfig[]>;
  }

  getPipelineFeed(opts?: Record<string, string | number | boolean>): Promise<PipelineFeed[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/user/feed?${query}`) as Promise<PipelineFeed[]>;
  }

  cancelPipeline(owner: string, repo: string, pipelineNumber: number): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/pipelines/${pipelineNumber}/cancel`);
  }

  approvePipeline(owner: string, repo: string, pipelineNumber: string): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/pipelines/${pipelineNumber}/approve`);
  }

  declinePipeline(owner: string, repo: string, pipelineNumber: string): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/pipelines/${pipelineNumber}/decline`);
  }

  skipPipelineWorkflow(owner: string, repo: string, pipeline: string, workflowId: number): Promise<unknown> {
    return this._post(`/api/repos/${owner}/${repo}/pipelines/${pipeline}/skip/${workflowId}`);
  }

  restartPipeline(
    owner: string,
    repo: string,
    pipeline: string,
    opts?: Record<string, string | number | boolean>,
  ): Promise<Pipeline> {
    const query = encodeQueryString(opts);
    return this._post(`/api/repos/${owner}/${repo}/pipelines/${pipeline}?${query}`) as Promise<Pipeline>;
  }

  getLogs(owner: string, repo: string, pipeline: number, step: number): Promise<PipelineLog[]> {
    return this._get(`/api/repos/${owner}/${repo}/logs/${pipeline}/${step}`) as Promise<PipelineLog[]>;
  }

  getArtifact(owner: string, repo: string, pipeline: string, step: string, file: string): Promise<unknown> {
    return this._get(`/api/repos/${owner}/${repo}/files/${pipeline}/${step}/${file}?raw=true`);
  }

  getArtifactList(owner: string, repo: string, pipeline: string): Promise<unknown> {
    return this._get(`/api/repos/${owner}/${repo}/files/${pipeline}`);
  }

  getSecretList(owner: string, repo: string, page: number): Promise<Secret[] | null> {
    return this._get(`/api/repos/${owner}/${repo}/secrets?page=${page}`) as Promise<Secret[] | null>;
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

  getRegistryList(owner: string, repo: string, page: number): Promise<Registry[] | null> {
    return this._get(`/api/repos/${owner}/${repo}/registry?page=${page}`) as Promise<Registry[] | null>;
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

  getCronList(owner: string, repo: string, page: number): Promise<Cron[] | null> {
    return this._get(`/api/repos/${owner}/${repo}/cron?page=${page}`) as Promise<Cron[] | null>;
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

  runCron(owner: string, repo: string, cronId: number): Promise<Pipeline> {
    return this._post(`/api/repos/${owner}/${repo}/cron/${cronId}`) as Promise<Pipeline>;
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
    owner: string,
    repo: string,
    pipeline: number,
    step: number,
    // eslint-disable-next-line promise/prefer-await-to-callbacks
    callback: (data: PipelineLog) => void,
  ): EventSource {
    return this._subscribe(`/stream/logs/${owner}/${repo}/${pipeline}/${step}`, callback, {
      reconnect: true,
    });
  }
}
