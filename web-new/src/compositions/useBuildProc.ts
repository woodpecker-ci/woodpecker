import { ref } from 'vue';
import { BuildProc, BuildLog } from '~/lib/api/types';
import useApiClient from './useApiClient';

const apiClient = useApiClient();

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

export default () => {
  const logs = ref<BuildLog[] | undefined>();
  const proc = ref<BuildProc>();
  let stream: EventSource | undefined;

  function onLogsUpdate(data: BuildLog) {
    if (data.proc === proc.value?.name) {
      logs.value = [...(logs.value || []), data];
    }
  }

  async function load(owner: string, repo: string, build: number, _proc: BuildProc) {
    unload();

    proc.value = _proc;

    try {
      logs.value = await apiClient.getLogs(owner, repo, build, _proc.pid);
    } catch (err) {
      logs.value = [];
    }

    if (isProcRunning(_proc)) {
      stream = apiClient.streamLogs(owner, repo, build, _proc.ppid, onLogsUpdate);
    }
  }

  function unload() {
    if (stream) {
      stream.close();
    }
  }

  return { logs, load, unload };
};
