<template>
  <AgentManager
    :description="$t('org.settings.agents.desc')"
    :load-agents="loadAgents"
    :create-agent="createAgent"
    :update-agent="updateAgent"
    :delete-agent="deleteAgent"
  >
    <template #headerEnd>
      <div class="mt-4">
        <Checkbox
          :model-value="org.require_org_agents"
          :label="$t('org.settings.agents.require_org_agents.require')"
          :description="$t('org.settings.agents.require_org_agents.desc')"
          :disabled="isSavingOrg"
          @update:model-value="saveRequireOrgAgents"
        />
      </div>
    </template>
  </AgentManager>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import AgentManager from '~/components/agent/AgentManager.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent } from '~/lib/api/types';

const apiClient = useApiClient();
const org = requiredInject('org');

const loadAgents = (page: number) => apiClient.getOrgAgents(org.value.id, { page });
const createAgent = (agent: Partial<Agent>) => apiClient.createOrgAgent(org.value.id, agent);
const updateAgent = (agent: Agent) => apiClient.updateOrgAgent(org.value.id, agent.id, agent);
const deleteAgent = (agent: Agent) => apiClient.deleteOrgAgent(org.value.id, agent.id);

const { t } = useI18n();
const notifications = useNotifications();

const { doSubmit: saveRequireOrgAgents, isLoading: isSavingOrg } = useAsyncAction(async (requireOrgAgents: boolean) => {
  try {
    org.value = await apiClient.updateOrg(org.value.id, { require_org_agents: requireOrgAgents });
    notifications.notify({ title: t('org.settings.agents.require_org_agents.saved'), type: 'success' });
  } catch (error) {
    notifications.notifyError(error as Error);
  }
});

useWPTitle(computed(() => [t('admin.settings.agents.agents'), org.value.name]));
</script>
