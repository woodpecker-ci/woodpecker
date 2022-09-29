import { Pipeline, PipelineProc, Repo } from '~/lib/api/types';

export function findProc(procs: PipelineProc[], pid: number): PipelineProc | undefined {
  return procs.reduce((prev, proc) => {
    if (proc.pid === pid) {
      return proc;
    }

    if (proc.children) {
      const result = findProc(proc.children, pid);
      if (result) {
        return result;
      }
    }

    return prev;
  }, undefined as PipelineProc | undefined);
}

/**
 * Returns true if the process is in a completed state.
 *
 * @param {Object} proc - The process object.
 * @returns {boolean}
 */
export function isProcFinished(proc: PipelineProc): boolean {
  return proc.state !== 'running' && proc.state !== 'pending';
}

/**
 * Returns true if the process is running.
 *
 * @param {Object} proc - The process object.
 * @returns {boolean}
 */
export function isProcRunning(proc: PipelineProc): boolean {
  return proc.state === 'running';
}

/**
 * Compare two pipelines by name.
 * @param {Object} a - A pipeline.
 * @param {Object} b - A pipeline.
 * @returns {number}
 */
export function comparePipelines(a: Pipeline, b: Pipeline): number {
  return (b.created_at || -1) - (a.created_at || -1);
}

export function isPipelineActive(pipeline: Pipeline): boolean {
  return ['pending', 'running', 'started'].includes(pipeline.status);
}

export function repoSlug(ownerOrRepo: Repo): string;
export function repoSlug(ownerOrRepo: string, name: string): string;
export function repoSlug(ownerOrRepo: string | Repo, name?: string): string {
  if (typeof ownerOrRepo === 'string') {
    if (!name) {
      throw new Error('Please provide a name as well');
    }

    return `${ownerOrRepo}/${name}`;
  }

  return `${ownerOrRepo.owner}/${ownerOrRepo.name}`;
}
