import type { Pipeline } from './pipeline';

// A version control repository.
export interface Repo {
  // Is the repo currently active or not
  active: boolean;

  // The unique identifier for the repository.
  id: number;

  // The id of the repository on the source control management system.
  forge_remote_id: string;

  // The id of the forge that the repository is on.
  forge_id: number;

  // The source control management being used.
  // Currently, this is either 'git' or 'hg' (Mercurial).
  scm: string;

  // Whether the forge repo has PRs enabled.
  pr_enabled: boolean;

  // The id of the organization that owns the repository.
  org_id: number;

  // The owner of the repository.
  owner: string;

  // The name of the repository.
  name: string;

  // The full name of the repository.
  // This is created from the owner and name of the repository.
  full_name: string;

  // The url for the avatar image.
  avatar_url: string;

  // The url to view the repository.
  forge_url: string;

  // The url used to clone the repository.
  clone_url: string;

  // The default branch of the repository.
  default_branch: string;

  // Whether the repository is publicly visible.
  private: boolean;

  // Whether the repository has trusted access for pipelines.
  // If the repository is trusted then the host network can be used and
  // volumes can be created.
  trusted: RepoTrusted;

  // x-dart-type: Duration
  // The amount of time in minutes before the pipeline is killed.
  timeout: number;

  // Whether pull requests should trigger a pipeline.
  allow_pr: boolean;

  allow_deploy: boolean;

  config_file: string;

  visibility: RepoVisibility;

  last_pipeline: Pipeline;

  require_approval: RepoRequireApproval;

  // Events that will cancel running pipelines before starting a new one
  cancel_previous_pipeline_events: string[];

  netrc_only_trusted: boolean;
}

/* eslint-disable no-unused-vars */
export enum RepoVisibility {
  Public = 'public',
  Private = 'private',
  Internal = 'internal',
}

export enum RepoRequireApproval {
  None = 'none',
  Forks = 'forks',
  PullRequests = 'pull_requests',
  AllEvents = 'all_events',
}
/* eslint-enable */

export type RepoSettings = Pick<
  Repo,
  | 'config_file'
  | 'timeout'
  | 'visibility'
  | 'trusted'
  | 'require_approval'
  | 'allow_pr'
  | 'allow_deploy'
  | 'cancel_previous_pipeline_events'
  | 'netrc_only_trusted'
>;

export interface RepoPermissions {
  pull: boolean;
  push: boolean;
  admin: boolean;
  synced: number;
}

export interface RepoTrusted {
  network: boolean;
  volumes: boolean;
  security: boolean;
}
