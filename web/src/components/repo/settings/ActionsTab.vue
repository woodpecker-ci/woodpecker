<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-gray-500">Actions</h1>
    </div>

    <div class="flex flex-col">
      <Button class="mr-auto mt-4" color="blue" text="Repair repository" @click="repairRepo" />

      <Button class="mr-auto mt-4" color="red" text="Delete repository" @click="deleteRepo" />
    </div>
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

    async function repairRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await apiClient.repairRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository repaired', type: 'success' });
    }

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
      repairRepo,
    };
  },
});
</script>
