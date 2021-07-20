<template>
  <div class="max-w-5xl w-full h-full m-auto">
    <div class="flex w-full border-b mb-4 py-2">
      <Button class="ml-auto" @click="reloadRepos" text="Reload Repos" />
    </div>
    <FluidContainer>
      <div v-for="repo in repos" :key="repo.id" class="flex mb-4">
        <span>{{ repo.owner }} / {{ repo.name }}</span>
        <Button v-if="!repo.active" class="ml-auto" @click="activateRepo(repo)" text="Activate" />
      </div>
    </FluidContainer>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { Repo } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import router from '~/router';

export default defineComponent({
  name: 'RepoAdd',

  components: {
    Button,
    FluidContainer,
  },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();
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
      notifications.notify({ title: 'Repository enabled', type: 'success' });
      await router.push({ name: 'repo', params: { repoId: repo.name, repoOwner: repo.owner } });
    }

    return { repos, reloadRepos, activateRepo };
  },
});
</script>
