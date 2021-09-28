import { Build, BuildProc, Repo } from '~/lib/api/types';

export function findProc(procs: BuildProc[], pid: number): BuildProc | undefined {
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
  }, undefined as BuildProc | undefined);
}

/**
 * Returns true if the process is in a completed state.
 *
 * @param {Object} proc - The process object.
 * @returns {boolean}
 */
export function isProcFinished(proc: BuildProc): boolean {
  return proc.state !== 'running' && proc.state !== 'pending';
}

/**
 * Returns true if the process is running.
 *
 * @param {Object} proc - The process object.
 * @returns {boolean}
 */
export function isProcRunning(proc: BuildProc): boolean {
  return proc.state === 'running';
}

/**
 * Compare two builds by name.
 * @param {Object} a - A build.
 * @param {Object} b - A build.
 * @returns {number}
 */
export function compareBuilds(a: Build, b: Build): number {
  return (b.started_at || b.created_at || -1) - (a.started_at || a.created_at || -1);
}

export function isBuildActive(build: Build): boolean {
  return ['pending', 'running', 'started'].includes(build.status);
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
