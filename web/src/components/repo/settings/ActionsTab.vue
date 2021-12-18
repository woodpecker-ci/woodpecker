<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-gray-500">Actions</h1>
    </div>

    <div class="flex flex-col">
      <Button
        class="mr-auto mt-4"
        color="blue"
        start-icon="heal"
        text="Repair repository"
        :is-loading="isRepairingRepo"
        @click="repairRepo"
      />

      <Button
        class="mr-auto mt-4"
        color="blue"
        start-icon="turn-off"
        text="Disable repository"
        :is-loading="isDeactivatingRepo"
        @click="deactivateRepo"
      />

      <Button
        class="mr-auto mt-4"
        color="red"
        start-icon="trash"
        text="Delete repository"
        :is-loading="isDeletingRepo"
        @click="deleteRepo"
      />
    </div>
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, Ref } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
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

    const { doSubmit: repairRepo, isLoading: isRepairingRepo } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await apiClient.repairRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository repaired', type: 'success' });
    });

    const { doSubmit: deleteRepo, isLoading: isDeletingRepo } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      // TODO use proper dialog
      // eslint-disable-next-line no-alert, no-restricted-globals
      if (!confirm('All data will be lost after this action!!!\n\nDo you really want to proceed?')) {
        return;
      }

      await apiClient.deleteRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository deleted', type: 'success' });
      await router.replace({ name: 'repos' });
    });

    const { doSubmit: deactivateRepo, isLoading: isDeactivatingRepo } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await apiClient.deleteRepo(repo.value.owner, repo.value.name, false);
      notifications.notify({ title: 'Repository disabled', type: 'success' });
      await router.replace({ name: 'repos' });
    });

    return {
      isRepairingRepo,
      isDeletingRepo,
      isDeactivatingRepo,
      deleteRepo,
      repairRepo,
      deactivateRepo,
    };
  },
});
</script>
