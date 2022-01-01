<template>
  <FluidContainer v-if="buildConfigs" class="flex flex-col gap-y-6 text-gray-500 justify-between py-0">
    <div v-for="buildConfig in buildConfigs" :key="buildConfig.hash" class="w-full">
      <div class="font-bold">{{ buildConfig.name }}</div>
      <div class="w-full bg-gray-400 dark:bg-dark-gray-700 rounded-md p-2 font-mono whitespace-pre">
        {{ buildConfig.data }}
      </div>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref, watch } from 'vue';

import FluidContainer from '~/components/layout/FluidContainer.vue';
import useApiClient from '~/compositions/useApiClient';
import { Build, BuildConfig, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildConfig',

  components: {
    FluidContainer,
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
