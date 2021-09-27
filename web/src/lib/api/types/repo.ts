// A version control repository.
export type Repo = {
  active: boolean;
  // Is the repo currently active or not

  id: number;
  // The unique identifier for the repository.

  scm: string;
  // The source control management being used.
  // Currently this is either 'git' or 'hg' (Mercurial).

  owner: string;
  // The owner of the repository.

  name: string;
  // The name of the repository.

  full_name: string;
  // The full name of the repository.
  // This is created from the owner and name of the repository.

  avatar_url: string;
  // The url for the avatar image.

  link_url: string;
  // The link to view the repository.

  clone_url: string;
  // The url used to clone the repository.

  default_branch: string;
  // The default branch of the repository.

  private: boolean;
  // Whether the repository is publicly visible.

  trusted: boolean;
  // Whether the repository has trusted access for builds.
  // If the repository is trusted then the host network can be used and
  // volumes can be created.

  timeout: number;
  // x-dart-type: Duration
  // The amount of time in minutes before the build is killed.

  allow_pr: boolean;
  // Whether pull requests should trigger a build.

  config_file: string;

  visibility: RepoVisibility;

  last_build: number;

  gated: boolean;
};

export enum RepoVisibility {
  Public = 'public',
  Private = 'private',
  Internal = 'internal',
}

export type RepoSettings = Pick<Repo, 'config_file' | 'timeout' | 'visibility' | 'trusted' | 'gated' | 'allow_pr'>;

export type RepoPermissions = {
  pull: boolean;
  push: boolean;
  admin: boolean;
  synced: number;
};
