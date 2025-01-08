import type { PullRequest } from './pull_request';
import type { WebhookEvents } from './webhook';

export interface PipelineError<D = unknown> {
  type: string;
  message: string;
  data?: D;
  is_warning: boolean;
}

// A pipeline for a repository.
export interface Pipeline {
  id: number;

  // The pipeline number.
  // This number is specified within the context of the repository the pipeline belongs to and is unique within that.
  number: number;

  parent: number;

  event: WebhookEvents;

  //  The current status of the pipeline.
  status: PipelineStatus;

  errors?: PipelineError[];

  // When the pipeline request was received.
  created: number;

  // When the pipeline was updated last time in database.
  updated: number;

  // When the pipeline began execution.
  started: number;

  // When the pipeline was finished.
  finished: number;

  // The commit for the pipeline.
  commit: PipelineCommit;

  // The branch the commit was pushed to.
  branch: string;

  // The alias for the commit.
  ref: string;

  // The mapping from the local repository to a branch in the forge.
  refspec: string;

  // The login for the author of the commit.
  author: string;

  // The avatar for the author of the commit.
  author_avatar: string;

  // This url will point to the repository state associated with the pipeline's commit.
  forge_url: string;

  reviewed_by: string;

  reviewed: number;

  pull_request?: PullRequest;

  deployment?: PipelineDeployment;

  release_title?: string;

  cron: string;

  // The steps associated with this pipeline.
  // A pipeline will have multiple steps if a matrix pipeline was used or if a rebuild was requested.
  workflows?: PipelineWorkflow[];

  changed_files?: string[];
}

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

export interface PipelineWorkflow {
  id: number;
  pipeline_id: number;
  pid: number;
  name: string;
  state: PipelineStatus;
  environ?: Record<string, string>;
  started?: number;
  finished?: number;
  agent_id?: number;
  error?: string;
  children: PipelineStep[];
}

export interface PipelineStep {
  id: number;
  uuid: string;
  pipeline_id: number;
  pid: number;
  ppid: number;
  name: string;
  state: PipelineStatus;
  exit_code: number;
  started?: number;
  finished?: number;
  error?: string;
  type?: StepType;
}

export interface PipelineCommit {
  sha: string;
  message: string;
  forge_url: string;
  author: PipelineCommitAuthor;
}

export interface PipelineCommitAuthor {
  author: string;
  email: string;
}

export interface PipelineDeployment {
  description: string;
}

export interface PipelineLog {
  id: number;
  step_id: number;
  time: number;
  line: number;
  data: string; // base64 encoded
  type: number;
}

export type PipelineFeed = Pipeline & {
  repo_id: number;
};

/* eslint-disable no-unused-vars */
export enum StepType {
  Clone = 'clone',
  Service = 'service',
  Plugin = 'plugin',
  Commands = 'commands',
  Cache = 'cache',
}
/* eslint-enable */
