/* eslint-disable no-unused-vars */
// A version control repository.
export interface Repo {
  // Is the repo currently active or not
  active: boolean;

  // The unique identifier for the repository.
  id: number;

  // The id of the repository on the source control management system.
  forge_remote_id: string;

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
  trusted: boolean;

  // x-dart-type: Duration
  // The amount of time in minutes before the pipeline is killed.
  timeout: number;

  // Whether pull requests should trigger a pipeline.
  allow_pr: boolean;

  allow_deploy: boolean;

  config_file: string;

  visibility: RepoVisibility;

  last_pipeline: number;

  gated: boolean;

  // Events that will cancel running pipelines before starting a new one
  cancel_previous_pipeline_events: string[];

  netrc_only_trusted: boolean;
}

export enum RepoVisibility {
  Public = 'public',
  Private = 'private',
  Internal = 'internal',
}

export type RepoSettings = Pick<
  Repo,
  | 'config_file'
  | 'timeout'
  | 'visibility'
  | 'trusted'
  | 'gated'
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
