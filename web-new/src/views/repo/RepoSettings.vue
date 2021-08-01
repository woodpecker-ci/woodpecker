<template>
  <FluidContainer v-if="repo">
    <div class="flex border-b items-center pb-4 mb-4">
      <!-- <Breadcrumbs
        :paths="[
          { name: 'Repositories', link: { name: 'home' } },
          {
            name: `${repo.owner} / ${repo.name}`,
            link: { name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } },
          },
          { name: 'Settings', link: { name: 'repo-settings', params: { repoOwner: repo.owner, repoName: repo.name } } },
        ]"
      /> -->
      <IconButton :to="{ name: 'repo' }">
        <IconBack class="w-8 h-8" />
      </IconButton>
      <h1 class="text-xl ml-2">Settings</h1>
    </div>

    <div>TODO</div>
  </FluidContainer>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, Ref, ref, toRef, watch } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import { useRouter } from 'vue-router';
import useNotifications from '~/compositions/useNotifications';
import Breadcrumbs from '~/components/layout/Breadcrumbs.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import IconBack from 'virtual:vite-icons/iconoir/arrow-left.vue';

export default defineComponent({
  name: 'RepoSettings',

  components: { FluidContainer, Button, Breadcrumbs, IconButton, IconBack },

  setup() {
    const apiClient = useApiClient();
    const router = useRouter();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');
    const badgeUrl = computed(() => {
      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      return `/api/badges/${repo.value.owner}/${repo.value.name}/status.svg`;
    });

    async function disableRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await apiClient.deleteRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository deleted', type: 'success' });
      await router.replace({ name: 'repos' });
    }

    return { repo, disableRepo, badgeUrl };
  },
});
</script>
