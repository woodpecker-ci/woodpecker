<template>
  <Settings :title="$t('repo.settings.actions.actions')">
    <div class="flex flex-wrap items-center">
      <Button
        class="mr-4 my-1"
        start-icon="heal"
        :is-loading="isRepairingRepo"
        :text="$t('repo.settings.actions.repair.repair')"
        @click="repairRepo"
      />

      <Button
        v-if="isActive"
        class="mr-4 my-1"
        start-icon="turn-off"
        :is-loading="isDeactivatingRepo"
        :text="$t('repo.settings.actions.disable.disable')"
        @click="deactivateRepo"
      />
      <Button
        v-else
        class="mr-4 my-1"
        color="blue"
        start-icon="turn-off"
        :is-loading="isActivatingRepo"
        :text="$t('repo.settings.actions.enable.enable')"
        @click="activateRepo"
      />

      <Button
        class="mr-4 my-1"
        color="red"
        start-icon="trash"
        :is-loading="isDeletingRepo"
        :text="$t('repo.settings.actions.delete.delete')"
        @click="deleteRepo"
      />
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed, inject, Ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { Repo } from '~/lib/api/types';

const apiClient = useApiClient();
const router = useRouter();
const notifications = useNotifications();
const i18n = useI18n();

const repo = inject<Ref<Repo>>('repo');

const { doSubmit: repairRepo, isLoading: isRepairingRepo } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo should be set');
  }

  await apiClient.repairRepo(repo.value.id);
  notifications.notify({ title: i18n.t('repo.settings.actions.repair.success'), type: 'success' });
});

const { doSubmit: deleteRepo, isLoading: isDeletingRepo } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo should be set');
  }

  // TODO use proper dialog
  // eslint-disable-next-line no-alert, no-restricted-globals
  if (!confirm(i18n.t('repo.settings.actions.delete.confirm'))) {
    return;
  }

  await apiClient.deleteRepo(repo.value.id);
  notifications.notify({ title: i18n.t('repo.settings.actions.delete.success'), type: 'success' });
  await router.replace({ name: 'repos' });
});

const { doSubmit: activateRepo, isLoading: isActivatingRepo } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo should be set');
  }

  await apiClient.activateRepo(repo.value.forge_remote_id);
  notifications.notify({ title: i18n.t('repo.settings.actions.enable.success'), type: 'success' });
});

const { doSubmit: deactivateRepo, isLoading: isDeactivatingRepo } = useAsyncAction(async () => {
  if (!repo) {
    throw new Error('Unexpected: Repo should be set');
  }

  await apiClient.deleteRepo(repo.value.id, false);
  notifications.notify({ title: i18n.t('repo.settings.actions.disable.success'), type: 'success' });
  await router.replace({ name: 'repos' });
});

const isActive = computed(() => repo?.value.active);
</script>
