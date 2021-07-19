<template>
  <div v-if="repo" class="max-w-5xl w-full h-full m-auto">
    <div class="flex border-b mb-4 p-4 lg:px-0 m-auto">
      <div>
        <span>{{ repo.owner }}</span> / <span>{{ repo.name }}</span>
      </div>
      <img class="ml-auto" :src="`/api/badges/${repoOwner}/${repoId}/status.svg`" />
    </div>
    <pre>{{ repo }}</pre>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref, toRef } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'Repo',

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
    repoId: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const apiClient = useApiClient();
    const repo = ref<Repo | undefined>();

    const repoOwner = toRef(props, 'repoOwner');
    const repoId = toRef(props, 'repoId');

    onMounted(async () => {
      repo.value = await apiClient.getRepo(repoOwner.value, repoId.value);
    });

    return { repo };
  },
});
</script>
