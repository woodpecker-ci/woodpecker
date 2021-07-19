<template>
  <div class="max-w-5xl w-full h-full m-auto">
    <div class="flex w-full border-b mb-4 py-2">
      <Button class="ml-auto" @click="reloadRepos" text="Reload Repos" />
    </div>
    <div v-for="repo in repos" :key="repo.id" class="flex mb-4">
      <span>{{ repo.owner }} / {{ repo.name }}</span>
      <Button v-if="repo.id" class="ml-auto" @click="activateRepo(repo)" text="Activate" />
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';

export default defineComponent({
  name: 'RepoAdd',

  components: {
    Button,
  },

  setup() {
    const apiClient = useApiClient();
    const repos = ref<Repo[] | undefined>();

    onMounted(async () => {
      repos.value = await apiClient.getRepoList({ all: true });
    });

    async function reloadRepos(): Promise<void> {
      repos.value = undefined;
      repos.value = await apiClient.getRepoList({ all: true, flush: true });
    }

    async function activateRepo(repo: Repo): Promise<void> {
      await apiClient.activateRepo(repo.owner, repo.name);
    }

    return { repos, reloadRepos, activateRepo };
  },
});
</script>
