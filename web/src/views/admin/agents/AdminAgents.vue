<template>
  <Settings :title="$t('admin.settings.agents.agents')" :description="$t('admin.settings.agents.desc')">
    <template #headerActions>
      <Button
        :text="$t('admin.settings.agents.add')"
        start-icon="plus"
        @click="$router.push({ name: 'admin-settings-agent-create' })"
      />
    </template>

    <AgentList
      :loading="loading"
      :agents="agents"
      :is-deleting="isDeleting"
      is-admin
      @edit="$router.push({ name: 'admin-settings-agent', params: { agentId: $event.id } })"
      @delete="deleteAgent"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import AgentList from '~/components/agent/AgentList.vue';
import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useInterval } from '~/compositions/useInterval';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent } from '~/lib/api/types';

const notifications = useNotifications();
const { t } = useI18n();
const apiClient = useApiClient();

const { resetPage, data: agents, loading } = usePagination((page: number) => apiClient.getAgents({ page }));

const { doSubmit: deleteAgent, isLoading: isDeleting } = useAsyncAction(async (agent: Agent) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('admin.settings.agents.delete_confirm'))) {
    return;
  }

  await apiClient.deleteAgent(agent);
  notifications.notify({ title: t('admin.settings.agents.deleted'), type: 'success' });
  await resetPage();
});

useInterval(resetPage, 5 * 1000); // refresh every 5s

useWPTitle(computed(() => [t('admin.settings.agents.agents'), t('admin.settings.settings')]));
</script>
