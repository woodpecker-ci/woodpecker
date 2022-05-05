// A version control repository.
export type Repo = {
  // Is the repo currently active or not
  active: boolean;

  // The unique identifier for the repository.
  id: number;

  // The source control management being used.
  // Currently this is either 'git' or 'hg' (Mercurial).
  scm: string;

  // The owner of the repository.
  owner: string;

  // The name of the repository.
  name: string;

  // The full name of the repository.
  // This is created from the owner and name of the repository.
  full_name: string;

  // The url for the avatar image.
  avatar_url: string;

  // The link to view the repository.
  link_url: string;

  // The url used to clone the repository.
  clone_url: string;

  // The default branch of the repository.
  default_branch: string;

  // Whether the repository is publicly visible.
  private: boolean;

  // Whether the repository has trusted access for builds.
  // If the repository is trusted then the host network can be used and
  // volumes can be created.
  trusted: boolean;

  // x-dart-type: Duration
  // The amount of time in minutes before the build is killed.
  timeout: number;

  // Whether pull requests should trigger a build.
  allow_pr: boolean;

  config_file: string;

  visibility: RepoVisibility;

  last_build: number;

  gated: boolean;

  // Events that will cancel running pipelines before starting a new one
  cancel_previous_build_events: string[];
};

export enum RepoVisibility {
  Public = 'public',
  Private = 'private',
  Internal = 'internal',
}

export type RepoSettings = Pick<
  Repo,
  'config_file' | 'timeout' | 'visibility' | 'trusted' | 'gated' | 'allow_pr' | 'cancel_previous_build_events'
>;

export type RepoPermissions = {
  pull: boolean;
  push: boolean;
  admin: boolean;
  synced: number;
};
