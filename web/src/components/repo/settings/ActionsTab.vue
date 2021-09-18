<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">Actions</h1>
    </div>

    <Button class="mr-auto mt-4 bg-red-500 hover:bg-red-400 text-white" text="Delete repository" @click="deleteRepo" />
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, Ref } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'ActionsTab',

  components: { Button, Panel },

  setup() {
    const apiClient = useApiClient();
    const router = useRouter();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');

    async function deleteRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      // TODO use proper dialog
      // eslint-disable-next-line no-alert, no-restricted-globals
      if (!confirm('All data will be lost after this action!!!\n\nDo you really want to procceed?')) {
        return;
      }

      await apiClient.deleteRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository deleted', type: 'success' });
      await router.replace({ name: 'repos' });
    }

    return {
      deleteRepo,
    };
  },
});
</script>
