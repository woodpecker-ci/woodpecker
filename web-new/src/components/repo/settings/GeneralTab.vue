<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">General</h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>
    </div>

    <div class="flex">
      <Button class="mx-auto bg-red-500 hover:bg-red-400 text-white" text="Delete repository" @click="deleteRepo" />
    </div>
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

    async function deleteRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      if (!confirm('All data will be lost after this action!!!\n\nDo you really want to procceed?')) {
        return;
      }

      await apiClient.deleteRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository deleted', type: 'success' });
      await router.replace({ name: 'repos' });
    }

    return { deleteRepo, badgeUrl };
  },
});
</script>
