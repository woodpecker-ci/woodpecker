import type { Ref } from 'vue';
import { computed, ref, watch } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import type { Pipeline } from '~/lib/api/types';
import type { ReusedWorkflow } from '~/lib/pipeline';
import { reusedWorkflows } from '~/lib/pipeline';

// Restart chains are short, this only guards against a broken parent link.
// Keep in sync with maxRestartChain on the server.
const maxAncestors = 50;

// useReusedWorkflows loads the ancestors of a partial restart and returns the
// workflows it took over from them.
export default function useReusedWorkflows(repoId: Ref<number>, pipeline: Ref<Pipeline | undefined>) {
  const apiClient = useApiClient();
  const ancestors = ref<Pipeline[]>([]);
  let loadID = 0;

  watch(
    () => [repoId.value, pipeline.value?.number, pipeline.value?.partial_restart, pipeline.value?.parent] as const,
    async ([id, , partialRestart, parent]) => {
      const currentLoad = ++loadID;
      const loaded: Pipeline[] = [];
      let next = partialRestart === true ? (parent ?? 0) : 0;
      try {
        while (next > 0 && loaded.length < maxAncestors) {
          const ancestor = await apiClient.getPipeline(id, next);
          loaded.push(ancestor);
          next = ancestor.partial_restart === true ? ancestor.parent : 0;
        }
      } catch (error) {
        // Show what could be loaded, e.g. if an ancestor got deleted.
        console.error('failed to load the parent pipeline of a partial restart', error);
      }
      // A newer pipeline was selected meanwhile.
      if (currentLoad === loadID) {
        ancestors.value = loaded;
      }
    },
    { immediate: true },
  );

  return computed<ReusedWorkflow[]>(() => (pipeline.value ? reusedWorkflows(pipeline.value, ancestors.value) : []));
}
