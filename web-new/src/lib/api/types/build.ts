// A build for a repository.
export type Build = {
  id: number;

  // The build number.
  // This number is specified within the context of the repository the build belongs to and is unique within that.
  number: number;

  //  The current status of the build.
  status: BuildStatus;

  // When the build request was received.
  created_at: number;

  // When the build was enqueued.
  enqueued_at: number;

  // When the build began execution.
  started_at: number;

  // When the build was finished.
  finished_at: number;

  //  description: Where the deployment should go.
  deploy_to: string;

  // description: The commit for the build.
  commit: string;

  // description: The branch the commit was pushed to.
  branch: string;

  // description: The commit message.
  message: string;

  // description: When the commit was created.
  timestamp: number;

  // description: The alias for the commit.
  ref: string;

  // description: The mapping from the local repository to a branch in the remote.
  refspec: string;

  // description: The remote repository.
  remote: string;

  // description: The login for the author of the commit.
  author: string;

  // description: The avatar for the author of the commit.
  author_avatar: string;

  // description: The email for the author of the commit.
  author_email: string;

  //   The link to view the repository.
  //   This link will point to the repository state associated with the build's commit.
  link_url: string;

  // The jobs associated with this build.
  // A build will have multiple jobs if a matrix build was used or if a rebuild was requested.
  jobs: Job[];
};

export type BuildStatus = {};
export type Job = {};
