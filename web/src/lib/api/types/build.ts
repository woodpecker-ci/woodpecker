// A build for a repository.
export type Build = {
  id: number;

  // The build number.
  // This number is specified within the context of the repository the build belongs to and is unique within that.
  number: number;

  parent: number;

  event: 'push' | 'tag' | 'pull_request' | 'deployment';

  //  The current status of the build.
  status: BuildStatus;

  error: string;

  // When the build request was received.
  created_at: number;

  // When the build was updated last time in database.
  updated_at: number;

  // When the build was enqueued.
  enqueued_at: number;

  // When the build began execution.
  started_at: number;

  // When the build was finished.
  finished_at: number;

  // Where the deployment should go.
  deploy_to: string;

  // The commit for the build.
  commit: string;

  // The branch the commit was pushed to.
  branch: string;

  // The commit message.
  message: string;

  // When the commit was created.
  timestamp: number;

  // The alias for the commit.
  ref: string;

  // The mapping from the local repository to a branch in the remote.
  refspec: string;

  // The remote repository.
  remote: string;

  title: string;

  sender: string;

  // The login for the author of the commit.
  author: string;

  // The avatar for the author of the commit.
  author_avatar: string;

  //  email for the author of the commit.
  author_email: string;

  // The link to view the repository.
  // This link will point to the repository state associated with the build's commit.
  link_url: string;

  signed: boolean;

  verified: boolean;

  reviewed_by: string;

  reviewed_at: number;

  // The jobs associated with this build.
  // A build will have multiple jobs if a matrix build was used or if a rebuild was requested.
  procs?: BuildProc[];

  changed_files?: string[];
};

export type BuildStatus =
  | 'blocked'
  | 'declined'
  | 'error'
  | 'failure'
  | 'killed'
  | 'pending'
  | 'running'
  | 'skipped'
  | 'started'
  | 'success';

export type BuildProc = {
  id: number;
  build_id: number;
  pid: number;
  ppid: number;
  pgid: number;
  name: string;
  state: BuildStatus;
  exit_code: number;
  environ?: Record<string, string>;
  start_time?: number;
  end_time?: number;
  machine?: string;
  error?: string;
  children?: BuildProc[];
};

export type BuildLog = {
  proc: string;
  pos: number;
  out: string;
  time?: number;
};

export type BuildFeed = Build & {
  owner: string;
  name: string;
  full_name: string;
};
