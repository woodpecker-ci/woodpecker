<template>
  <div v-if="repo">
    <FluidContainer class="flex border-b mb-4 items-start items-center">
      <Breadcrumbs
        :paths="[
          repo.owner,
          { name: repo.name, link: { name: 'repo', params: { repoOwner: repo.owner, repoId: repo.name } } },
          { name: 'Settings', link: { name: 'repo-settings', params: { repoOwner: repo.owner, repoId: repo.name } } },
        ]"
      />
    </FluidContainer>
    <FluidContainer class="space-y-2">
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>

      <Button class="ml-4" text="Deactivate repository" @click="disableRepo" />
    </FluidContainer>
  </div>
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

export default defineComponent({
  name: 'RepoSettings',

  components: { FluidContainer, Button, Breadcrumbs },

  setup(props) {
    const apiClient = useApiClient();
    const router = useRouter();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');
    const badgeUrl = computed(() => {
      if (!repo) {
        return null;
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
