<template>
  <AgentManager
    :desc="$t('user.settings.agents.desc')"
    :load-agents="loadAgents"
    :create-agent="createAgent"
    :update-agent="updateAgent"
    :delete-agent="deleteAgent"
  />
</template>

<script lang="ts" setup>
import AgentManager from '~/components/agent/AgentManager.vue';
import useApiClient from '~/compositions/useApiClient';
import useAuthentication from '~/compositions/useAuthentication';
import type { Agent } from '~/lib/api/types';

const apiClient = useApiClient();
const { user } = useAuthentication();

if (!user) {
  throw new Error('Unexpected: User should be authenticated');
}

const loadAgents = (page: number) => apiClient.getOrgAgents(user.org_id, { page });
const createAgent = (agent: Partial<Agent>) => apiClient.createOrgAgent(user.org_id, agent);
const updateAgent = (agent: Agent) => apiClient.updateOrgAgent(user.org_id, agent.id, agent);
const deleteAgent = (agent: Agent) => apiClient.deleteOrgAgent(user.org_id, agent.id);
</script>
