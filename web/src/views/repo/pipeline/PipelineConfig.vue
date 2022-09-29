<template>
  <FluidContainer v-if="buildConfigs" class="flex flex-col gap-y-6 text-color justify-between !pt-0">
    <Panel v-for="buildConfig in buildConfigs" :key="buildConfig.hash" :title="buildConfig.name">
      <SyntaxHighlight class="font-mono whitespace-pre overflow-auto" language="yaml" :code="buildConfig.data" />
    </Panel>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref, watch } from 'vue';

import SyntaxHighlight from '~/components/atomic/SyntaxHighlight';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { Pipeline, PipelineConfig, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildConfig',

  components: {
    FluidContainer,
    Panel,
    SyntaxHighlight,
  },

  setup() {
    const build = inject<Ref<Pipeline>>('build');
    const apiClient = useApiClient();
    const repo = inject<Ref<Repo>>('repo');
    if (!repo || !build) {
      throw new Error('Unexpected: "repo" & "build" should be provided at this place');
    }

    const buildConfigs = ref<PipelineConfig[]>();
    async function loadBuildConfig() {
      if (!repo || !build) {
        throw new Error('Unexpected: "repo" & "build" should be provided at this place');
      }

      buildConfigs.value = (await apiClient.getPipelineConfig(repo.value.owner, repo.value.name, build.value.number)).map(
        (i) => ({
          ...i,
          data: atob(i.data),
        }),
      );
    }

    onMounted(() => {
      loadBuildConfig();
    });

    watch(build, () => {
      loadBuildConfig();
    });

    return { buildConfigs };
  },
});
</script>
