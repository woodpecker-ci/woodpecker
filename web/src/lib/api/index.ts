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
  async getRepoList(opts?: RepoListOptions): Promise<Repo[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/user/repos?${query}`) as Promise<Repo[]>;
  }

  async lookupRepo(owner: string, name: string): Promise<Repo | undefined> {
    return this._get(`/api/repos/lookup/${owner}/${name}`) as Promise<Repo | undefined>;
  }

  async getRepo(repoId: number): Promise<Repo> {
    return this._get(`/api/repos/${repoId}`) as Promise<Repo>;
  }

  async getRepoPermissions(repoId: number): Promise<RepoPermissions> {
    return this._get(`/api/repos/${repoId}/permissions`) as Promise<RepoPermissions>;
  }

  async getRepoBranches(repoId: number, opts?: PaginationOptions): Promise<string[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/branches?${query}`) as Promise<string[]>;
  }

  async getRepoPullRequests(repoId: number, opts?: PaginationOptions): Promise<PullRequest[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/pull_requests?${query}`) as Promise<PullRequest[]>;
  }

  async activateRepo(forgeRemoteId: string): Promise<Repo> {
    return this._post(`/api/repos?forge_remote_id=${forgeRemoteId}`) as Promise<Repo>;
  }

  async updateRepo(repoId: number, repoSettings: RepoSettings): Promise<unknown> {
    return this._patch(`/api/repos/${repoId}`, repoSettings);
  }

  async deleteRepo(repoId: number, remove = true): Promise<unknown> {
    const query = encodeQueryString({ remove });
    return this._delete(`/api/repos/${repoId}?${query}`);
  }

  async repairRepo(repoId: number): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/repair`);
  }

  async createPipeline(repoId: number, options: PipelineOptions): Promise<Pipeline> {
    return this._post(`/api/repos/${repoId}/pipelines`, options) as Promise<Pipeline>;
  }

  // Deploy triggers a deployment for an existing pipeline using the
  // specified target environment and task.
  async deployPipeline(repoId: number, pipelineNumber: string, options: DeploymentOptions): Promise<Pipeline> {
    const vars = {
      ...options.variables,
      event: 'deployment',
      deploy_to: options.environment,
      deploy_task: options.task,
    };
    const query = encodeQueryString(vars);
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}?${query}`) as Promise<Pipeline>;
  }

  async getPipelineList(
    repoId: number,
    opts?: PaginationOptions & { before?: string; after?: string },
  ): Promise<Pipeline[]> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/pipelines?${query}`) as Promise<Pipeline[]>;
  }

  async getPipeline(repoId: number, pipelineNumber: number | 'latest'): Promise<Pipeline> {
    return this._get(`/api/repos/${repoId}/pipelines/${pipelineNumber}`) as Promise<Pipeline>;
  }

  async getPipelineConfig(repoId: number, pipelineNumber: number): Promise<PipelineConfig[]> {
    return this._get(`/api/repos/${repoId}/pipelines/${pipelineNumber}/config`) as Promise<PipelineConfig[]>;
  }

  async getPipelineMetadata(repoId: number, pipelineNumber: number): Promise<any> {
    return this._get(`/api/repos/${repoId}/pipelines/${pipelineNumber}/metadata`) as Promise<any>;
  }

  async getPipelineFeed(): Promise<PipelineFeed[]> {
    return this._get(`/api/user/feed`) as Promise<PipelineFeed[]>;
  }

  async cancelPipeline(repoId: number, pipelineNumber: number): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}/cancel`);
  }

  async approvePipeline(repoId: number, pipelineNumber: string): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}/approve`);
  }

  async declinePipeline(repoId: number, pipelineNumber: string): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/pipelines/${pipelineNumber}/decline`);
  }

  async restartPipeline(
    repoId: number,
    pipeline: string,
    opts?: { event?: string; deploy_to?: string; fork?: boolean },
  ): Promise<Pipeline> {
    const query = encodeQueryString(opts);
    return this._post(`/api/repos/${repoId}/pipelines/${pipeline}?${query}`) as Promise<Pipeline>;
  }

  async getLogs(repoId: number, pipeline: number, step: number): Promise<PipelineLog[]> {
    return this._get(`/api/repos/${repoId}/logs/${pipeline}/${step}`) as Promise<PipelineLog[]>;
  }

  async deleteLogs(repoId: number, pipeline: number, step: number): Promise<unknown> {
    return this._delete(`/api/repos/${repoId}/logs/${pipeline}/${step}`);
  }

  async getSecretList(repoId: number, opts?: PaginationOptions): Promise<Secret[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/secrets?${query}`) as Promise<Secret[] | null>;
  }

  async createSecret(repoId: number, secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/secrets`, secret);
  }

  async updateSecret(repoId: number, secret: Partial<Secret>): Promise<unknown> {
    const secretName = encodeURIComponent(secret.name ?? '');
    return this._patch(`/api/repos/${repoId}/secrets/${secretName}`, secret);
  }

  async deleteSecret(repoId: number, secretName: string): Promise<unknown> {
    const name = encodeURIComponent(secretName);
    return this._delete(`/api/repos/${repoId}/secrets/${name}`);
  }

  async getRegistryList(repoId: number, opts?: PaginationOptions): Promise<Registry[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/registries?${query}`) as Promise<Registry[] | null>;
  }

  async createRegistry(repoId: number, registry: Partial<Registry>): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/registries`, registry);
  }

  async updateRegistry(repoId: number, registry: Partial<Registry>): Promise<unknown> {
    return this._patch(`/api/repos/${repoId}/registries/${registry.address}`, registry);
  }

  async deleteRegistry(repoId: number, registryAddress: string): Promise<unknown> {
    return this._delete(`/api/repos/${repoId}/registries/${registryAddress}`);
  }

  async getOrgRegistryList(orgId: number, opts?: PaginationOptions): Promise<Registry[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/orgs/${orgId}/registries?${query}`) as Promise<Registry[] | null>;
  }

  async createOrgRegistry(orgId: number, registry: Partial<Registry>): Promise<unknown> {
    return this._post(`/api/orgs/${orgId}/registries`, registry);
  }

  async updateOrgRegistry(orgId: number, registry: Partial<Registry>): Promise<unknown> {
    return this._patch(`/api/orgs/${orgId}/registries/${registry.address}`, registry);
  }

  async deleteOrgRegistry(orgId: number, registryAddress: string): Promise<unknown> {
    return this._delete(`/api/orgs/${orgId}/registries/${registryAddress}`);
  }

  async getGlobalRegistryList(opts?: PaginationOptions): Promise<Registry[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/registries?${query}`) as Promise<Registry[] | null>;
  }

  async createGlobalRegistry(registry: Partial<Registry>): Promise<unknown> {
    return this._post(`/api/registries`, registry);
  }

  async updateGlobalRegistry(registry: Partial<Registry>): Promise<unknown> {
    return this._patch(`/api/registries/${registry.address}`, registry);
  }

  async deleteGlobalRegistry(registryAddress: string): Promise<unknown> {
    return this._delete(`/api/registries/${registryAddress}`);
  }

  async getCronList(repoId: number, opts?: PaginationOptions): Promise<Cron[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos/${repoId}/cron?${query}`) as Promise<Cron[] | null>;
  }

  async createCron(repoId: number, cron: Partial<Cron>): Promise<unknown> {
    return this._post(`/api/repos/${repoId}/cron`, cron);
  }

  async updateCron(repoId: number, cron: Partial<Cron>): Promise<unknown> {
    return this._patch(`/api/repos/${repoId}/cron/${cron.id}`, cron);
  }

  async deleteCron(repoId: number, cronId: number): Promise<unknown> {
    return this._delete(`/api/repos/${repoId}/cron/${cronId}`);
  }

  async runCron(repoId: number, cronId: number): Promise<Pipeline> {
    return this._post(`/api/repos/${repoId}/cron/${cronId}`) as Promise<Pipeline>;
  }

  async getOrg(orgId: number): Promise<Org> {
    return this._get(`/api/orgs/${orgId}`) as Promise<Org>;
  }

  async lookupOrg(name: string): Promise<Org> {
    return this._get(`/api/orgs/lookup/${name}`) as Promise<Org>;
  }

  async getOrgPermissions(orgId: number): Promise<OrgPermissions> {
    return this._get(`/api/orgs/${orgId}/permissions`) as Promise<OrgPermissions>;
  }

  async getOrgSecretList(orgId: number, opts?: PaginationOptions): Promise<Secret[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/orgs/${orgId}/secrets?${query}`) as Promise<Secret[] | null>;
  }

  async createOrgSecret(orgId: number, secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/orgs/${orgId}/secrets`, secret);
  }

  async updateOrgSecret(orgId: number, secret: Partial<Secret>): Promise<unknown> {
    const secretName = encodeURIComponent(secret.name ?? '');
    return this._patch(`/api/orgs/${orgId}/secrets/${secretName}`, secret);
  }

  async deleteOrgSecret(orgId: number, secretName: string): Promise<unknown> {
    const name = encodeURIComponent(secretName);
    return this._delete(`/api/orgs/${orgId}/secrets/${name}`);
  }

  async getGlobalSecretList(opts?: PaginationOptions): Promise<Secret[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/secrets?${query}`) as Promise<Secret[] | null>;
  }

  async createGlobalSecret(secret: Partial<Secret>): Promise<unknown> {
    return this._post(`/api/secrets`, secret);
  }

  async updateGlobalSecret(secret: Partial<Secret>): Promise<unknown> {
    const secretName = encodeURIComponent(secret.name ?? '');
    return this._patch(`/api/secrets/${secretName}`, secret);
  }

  async deleteGlobalSecret(secretName: string): Promise<unknown> {
    const name = encodeURIComponent(secretName);
    return this._delete(`/api/secrets/${name}`);
  }

  async getSelf(): Promise<unknown> {
    return this._get('/api/user');
  }

  async getToken(): Promise<string> {
    return this._post('/api/user/token') as Promise<string>;
  }

  async getAgents(opts?: PaginationOptions): Promise<Agent[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/agents?${query}`) as Promise<Agent[] | null>;
  }

  async getAgent(agentId: Agent['id']): Promise<Agent> {
    return this._get(`/api/agents/${agentId}`) as Promise<Agent>;
  }

  async createAgent(agent: Partial<Agent>): Promise<Agent> {
    return this._post('/api/agents', agent) as Promise<Agent>;
  }

  async updateAgent(agent: Partial<Agent>): Promise<Agent> {
    return this._patch(`/api/agents/${agent.id}`, agent) as Promise<Agent>;
  }

  async deleteAgent(agent: Agent): Promise<unknown> {
    return this._delete(`/api/agents/${agent.id}`);
  }

  async getOrgAgents(orgId: number, opts?: PaginationOptions): Promise<Agent[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/orgs/${orgId}/agents?${query}`) as Promise<Agent[] | null>;
  }

  async createOrgAgent(orgId: number, agent: Partial<Agent>): Promise<Agent> {
    return this._post(`/api/orgs/${orgId}/agents`, agent) as Promise<Agent>;
  }

  async updateOrgAgent(orgId: number, agentId: number, agent: Partial<Agent>): Promise<Agent> {
    return this._patch(`/api/orgs/${orgId}/agents/${agentId}`, agent) as Promise<Agent>;
  }

  async deleteOrgAgent(orgId: number, agentId: number): Promise<unknown> {
    return this._delete(`/api/orgs/${orgId}/agents/${agentId}`);
  }

  async getForges(opts?: PaginationOptions): Promise<Forge[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/forges?${query}`) as Promise<Forge[] | null>;
  }

  async getForge(forgeId: Forge['id']): Promise<Forge> {
    return this._get(`/api/forges/${forgeId}`) as Promise<Forge>;
  }

  async createForge(forge: Partial<Forge>): Promise<Forge> {
    return this._post('/api/forges', forge) as Promise<Forge>;
  }

  async updateForge(forge: Partial<Forge>): Promise<unknown> {
    return this._patch(`/api/forges/${forge.id}`, forge);
  }

  async deleteForge(forge: Forge): Promise<unknown> {
    return this._delete(`/api/forges/${forge.id}`);
  }

  async getQueueInfo(): Promise<QueueInfo> {
    return this._get('/api/queue/info') as Promise<QueueInfo>;
  }

  async pauseQueue(): Promise<unknown> {
    return this._post('/api/queue/pause');
  }

  async resumeQueue(): Promise<unknown> {
    return this._post('/api/queue/resume');
  }

  async getUsers(opts?: PaginationOptions): Promise<User[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/users?${query}`) as Promise<User[] | null>;
  }

  async getUser(username: string): Promise<User> {
    return this._get(`/api/users/${username}`) as Promise<User>;
  }

  async createUser(user: Partial<User>): Promise<User> {
    return this._post('/api/users', user) as Promise<User>;
  }

  async updateUser(user: Partial<User>): Promise<unknown> {
    return this._patch(`/api/users/${user.login}`, user);
  }

  async deleteUser(user: User): Promise<unknown> {
    return this._delete(`/api/users/${user.login}`);
  }

  async resetToken(): Promise<string> {
    return this._delete('/api/user/token') as Promise<string>;
  }

  async getOrgs(opts?: PaginationOptions): Promise<Org[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/orgs?${query}`) as Promise<Org[] | null>;
  }

  async deleteOrg(org: Org): Promise<unknown> {
    return this._delete(`/api/orgs/${org.id}`);
  }

  async getAllRepos(opts?: PaginationOptions): Promise<Repo[] | null> {
    const query = encodeQueryString(opts);
    return this._get(`/api/repos?${query}`) as Promise<Repo[] | null>;
  }

  async repairAllRepos(): Promise<unknown> {
    return this._post(`/api/repos/repair`);
  }

  // eslint-disable-next-line promise/prefer-await-to-callbacks
  on(callback: (data: { pipeline?: Pipeline; repo?: Repo }) => void): EventSource {
    return this._subscribe('/api/stream/events', callback, {
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
