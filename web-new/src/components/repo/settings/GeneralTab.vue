<template>
  <Panel>
    <span class="text-lg border-b-2 w-full">General</span>
    <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
      <img :src="badgeUrl" />
    </a>

    <Button class="ml-4" text="Deactivate repository" @click="disableRepo" />
  </Panel>
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';
import { useRouter } from 'vue-router';
import useNotifications from '~/compositions/useNotifications';
import Panel from '~/components/layout/Panel.vue';

export default defineComponent({
  name: 'GeneralTab',

  components: { Button, Panel },

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
