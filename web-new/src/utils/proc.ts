import { BuildProc } from '~/lib/api/types';

export function findProc(procs: BuildProc[], pid: number): BuildProc | null {
  for (const proc of procs) {
    if (proc.pid === pid) {
      return proc;
    }
    if (proc.children) {
      const result = findProc(proc.children, pid);
      if (result) {
        return result;
      }
    }
  }

  return null;
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
