import { ref } from 'vue';

import { BuildLog, BuildProc } from '~/lib/api/types';
import { isProcFinished, isProcRunning } from '~/utils/helpers';

import useApiClient from './useApiClient';

const apiClient = useApiClient();

export default () => {
  const logs = ref<BuildLog[] | undefined>();
  const proc = ref<BuildProc>();
  let stream: EventSource | undefined;

  function onLogsUpdate(data: BuildLog) {
    if (data.proc === proc.value?.name) {
      logs.value = [...(logs.value || []), data];
    }
  }

  function unload() {
    if (stream) {
      stream.close();
    }
  }

  async function load(owner: string, repo: string, build: number, _proc: BuildProc) {
    unload();

    proc.value = _proc;
    logs.value = [];

    // we do not have logs for skipped jobs
    if (_proc.state === 'skipped' || _proc.state === 'killed') {
      return;
    }

    if (isProcFinished(_proc)) {
      logs.value = await apiClient.getLogs(owner, repo, build, _proc.pid);
    }

    if (isProcRunning(_proc)) {
      // load stream of parent process (which gets all processes logs)
      stream = apiClient.streamLogs(owner, repo, build, _proc.ppid, onLogsUpdate);
    }
  }

  return { logs, load, unload };
};
