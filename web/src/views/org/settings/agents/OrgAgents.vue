<template>
  <Settings :title="$t('agents')" :description="$t('agents_org_desc')">
    <template #headerActions>
      <Button :text="$t('add_agent')" start-icon="plus" :to="{ name: 'org-settings-agent-create' }" />
    </template>

    <AgentList
      :loading="loading"
      :agents="agents"
      :is-deleting="isDeleting"
      is-admin
      @edit="$router.push({ name: 'org-settings-agent', params: { agentId: $event.id } })"
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
import { requiredInject } from '~/compositions/useInjectProvide';
import { useInterval } from '~/compositions/useInterval';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent } from '~/lib/api/types';

const notifications = useNotifications();
const { t } = useI18n();
const apiClient = useApiClient();

const org = requiredInject('org');

const {
  resetPage,
  data: agents,
  loading,
} = usePagination((page: number) => apiClient.getOrgAgents(org.value.id, { page }));

const { doSubmit: deleteAgent, isLoading: isDeleting } = useAsyncAction(async (agent: Agent) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('agent_delete_confirm'))) {
    return;
  }

  await apiClient.deleteOrgAgent(org.value.id, agent);
  notifications.notify({ title: t('agent_deleted'), type: 'success' });
  await resetPage();
});

useInterval(resetPage, 5 * 1000, { immediate: false });

useWPTitle(computed(() => [t('agents'), t('settings')]));
</script>
