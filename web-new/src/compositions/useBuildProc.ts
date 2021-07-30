import { ref } from 'vue';
import { BuildProc, BuildLog } from '~/lib/api/types';
import { isProcRunning } from '~/utils/proc';
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
