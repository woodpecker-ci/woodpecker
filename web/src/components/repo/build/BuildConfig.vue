<template>
  <div v-if="buildConfig" class="flex mt-4 w-full bg-gray-600 min-h-0 flex-grow">
    <div class="flex flex-col w-3/12 text-white">
      <div v-for="config in buildConfig" :key="config.hash" @click="selectedConfigHash = config.hash">
        <div class="px-4 py-2 cursor-pointer hover:bg-gray-700">{{ config.name }}</div>
      </div>
    </div>

    <div class="w-9/12 flex-grow">
      <pre v-if="selectedConfig" class="text-gray-50 p-4">{{ selectedConfig.data }}</pre>
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, PropType, Ref, ref, toRef, watch } from 'vue';

import useApiClient from '~/compositions/useApiClient';
import { Build, BuildConfig, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'BuildConfig',

  props: {
    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    build: {
      type: Object as PropType<Build>,
      required: true,
    },
  },

  setup(props) {
    const apiClient = useApiClient();

    const build = toRef(props, 'build');
    const repo = inject<Ref<Repo>>('repo');
    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    const buildConfig = ref<BuildConfig[]>();
    const selectedConfigHash = ref<string>();
    const selectedConfig = computed(() => {
      if (!selectedConfigHash.value) {
        return undefined;
      }

      return buildConfig.value?.find((c) => c.hash === selectedConfigHash.value);
    });

    async function loadBuild(): Promise<void> {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      buildConfig.value = await apiClient.getBuildConfig(repo.value.owner, repo.value.name, build.value.number);
    }

    onMounted(loadBuild);
    watch([repo, build], loadBuild);

    return { buildConfig, selectedConfigHash, selectedConfig };
  },
});
</script>
