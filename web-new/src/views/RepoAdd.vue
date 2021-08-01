<template>
  <FluidContainer class="flex flex-col">
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <IconButton :to="{ name: 'repos' }">
        <IconBack class="w-8 h-8" />
      </IconButton>
      <h1 class="text-xl ml-2">Enable repository</h1>
      <Button class="ml-auto" @click="reloadRepos" text="Reload Repositories" />
    </div>

    <div class="space-y-4">
      <ListItem
        v-for="repo in repos"
        :key="repo.id"
        class="items-center"
        :clickable="repo.active"
        @click="repo.active && $router.push({ name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } })"
      >
        <span>{{ repo.full_name }}</span>
        <span v-if="repo.active" class="ml-auto">Already enabled</span>
        <Button v-if="!repo.active" class="ml-auto" @click="activateRepo(repo)" text="Activate" />
      </ListItem>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { Repo } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import router from '~/router';
import ListItem from '~/components/atomic/ListItem.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import IconBack from 'virtual:vite-icons/iconoir/arrow-left.vue';

export default defineComponent({
  name: 'RepoAdd',

  components: {
    Button,
    FluidContainer,
    ListItem,
    IconButton,
    IconBack,
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
      await router.push({ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } });
    }

    return { repos, reloadRepos, activateRepo };
  },
});
</script>
