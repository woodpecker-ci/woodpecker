<template>
  <AgentManager
    :desc="$t('admin.settings.agents.desc')"
    :load-agents="loadAgents"
    :create-agent="createAgent"
    :update-agent="updateAgent"
    :delete-agent="deleteAgent"
    :is-admin="true"
  />
</template>

<script lang="ts" setup>
import AgentManager from '~/components/agent/AgentManager.vue';
import useApiClient from '~/compositions/useApiClient';
import type { Agent } from '~/lib/api/types';

const apiClient = useApiClient();

const loadAgents = (page: number) => apiClient.getAgents({ page });
const createAgent = (agent: Partial<Agent>) => apiClient.createAgent(agent);
const updateAgent = (agent: Agent) => apiClient.updateAgent(agent);
const deleteAgent = (agent: Agent) => apiClient.deleteAgent(agent);
</script>
