// A pipeline for a repository.
export type Pipeline = {
  id: number;

  // The pipeline number.
  // This number is specified within the context of the repository the pipeline belongs to and is unique within that.
  number: number;

  parent: number;

  event: 'push' | 'tag' | 'pull_request' | 'deployment' | 'cron' | 'manual';

  //  The current status of the pipeline.
  status: PipelineStatus;

  error: string;

  // When the pipeline request was received.
  created_at: number;

  // When the pipeline was updated last time in database.
  updated_at: number;

  // When the pipeline was enqueued.
  enqueued_at: number;

  // When the pipeline began execution.
  started_at: number;

  // When the pipeline was finished.
  finished_at: number;

  // Where the deployment should go.
  deploy_to: string;

  // The commit for the pipeline.
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
  // This link will point to the repository state associated with the pipeline's commit.
  link_url: string;

  signed: boolean;

  verified: boolean;

  reviewed_by: string;

  reviewed_at: number;

  // The steps associated with this pipeline.
  // A pipeline will have multiple steps if a matrix pipeline was used or if a rebuild was requested.
  steps?: PipelineStep[];

  changed_files?: string[];
};

export type PipelineStatus =
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

export type PipelineStep = {
  id: number;
  pipeline_id: number;
  pid: number;
  ppid: number;
  pgid: number;
  name: string;
  state: PipelineStatus;
  exit_code: number;
  environ?: Record<string, string>;
  start_time?: number;
  end_time?: number;
  machine?: string;
  error?: string;
  children?: PipelineStep[];
};

export type PipelineLog = {
  step: string;
  pos: number;
  out: string;
  time?: number;
};

export type PipelineFeed = Pipeline & {
  owner: string;
  name: string;
  full_name: string;
};
