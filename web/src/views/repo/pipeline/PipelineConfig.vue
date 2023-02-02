<template>
  <div class="flex flex-col gap-y-6">
    <Panel v-for="pipelineConfig in pipelineConfigs || []" :key="pipelineConfig.hash" :title="pipelineConfig.name">
      <SyntaxHighlight class="font-mono whitespace-pre overflow-auto" language="yaml" :code="pipelineConfig.data" />
    </Panel>
  </div>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref, watch } from 'vue';

import SyntaxHighlight from '~/components/atomic/SyntaxHighlight';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { Pipeline, PipelineConfig, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'PipelineConfig',

  components: {
    Panel,
    SyntaxHighlight,
  },

  setup() {
    const pipeline = inject<Ref<Pipeline>>('pipeline');
    const apiClient = useApiClient();
    const repo = inject<Ref<Repo>>('repo');
    if (!repo || !pipeline) {
      throw new Error('Unexpected: "repo" & "pipeline" should be provided at this place');
    }

    const pipelineConfigs = ref<PipelineConfig[]>();
    async function loadPipelineConfig() {
      if (!repo || !pipeline) {
        throw new Error('Unexpected: "repo" & "pipeline" should be provided at this place');
      }

      pipelineConfigs.value = (
        await apiClient.getPipelineConfig(repo.value.owner, repo.value.name, pipeline.value.number)
      ).map((i) => ({
        ...i,
        data: atob(i.data),
      }));
    }

    onMounted(() => {
      loadPipelineConfig();
    });

    watch(pipeline, () => {
      loadPipelineConfig();
    });

    return { pipelineConfigs };
  },
});
</script>
