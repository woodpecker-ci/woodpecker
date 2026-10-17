// A version control pull request.
export interface PullRequest {
  // The index of the pull request.
  index: string;
  // The title of the pull request.
  title: string;
  // The branch the changes come from (empty if the forge does not provide it).
  source_branch?: string;
  // The branch the pull request will be merged into (empty if the forge does not provide it).
  target_branch?: string;
}
