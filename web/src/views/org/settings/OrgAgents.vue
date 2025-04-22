<template>
  <AgentManager
    :description="$t('org.settings.agents.desc')"
    :load-agents="loadAgents"
    :create-agent="createAgent"
    :update-agent="updateAgent"
    :delete-agent="deleteAgent"
  />
</template>

<script lang="ts" setup>
import AgentManager from '~/components/agent/AgentManager.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import type { Agent } from '~/lib/api/types';

const apiClient = useApiClient();
const org = requiredInject('org');

const loadAgents = (page: number) => apiClient.getOrgAgents(org.value.id, { page });
const createAgent = (agent: Partial<Agent>) => apiClient.createOrgAgent(org.value.id, agent);
const updateAgent = (agent: Agent) => apiClient.updateOrgAgent(org.value.id, agent.id, agent);
const deleteAgent = (agent: Agent) => apiClient.deleteOrgAgent(org.value.id, agent.id);
</script>
