<template>
  <FluidContainer v-if="buildConfigs" class="flex flex-col gap-y-6 text-gray-500 justify-between !py-0">
    <Panel v-for="buildConfig in buildConfigs" :key="buildConfig.hash" :title="buildConfig.name">
      <span class="font-mono whitespace-pre">{{ buildConfig.data }}</span>
    </Panel>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref, watch } from 'vue';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { Build, BuildConfig, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildConfig',

  components: {
    FluidContainer,
    Panel,
  },

  setup() {
    const build = inject<Ref<Build>>('build');
    const apiClient = useApiClient();
    const repo = inject<Ref<Repo>>('repo');
    if (!repo || !build) {
      throw new Error('Unexpected: "repo" & "build" should be provided at this place');
    }

    const buildConfigs = ref<BuildConfig[]>();
    async function loadBuildConfig() {
      if (!repo || !build) {
        throw new Error('Unexpected: "repo" & "build" should be provided at this place');
      }

      buildConfigs.value = (await apiClient.getBuildConfig(repo.value.owner, repo.value.name, build.value.number)).map(
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
