<template>
  <AgentManager
    :desc="$t('repo.settings.agents.desc')"
    :load-agents="loadAgents"
    :create-agent="createAgent"
    :update-agent="updateAgent"
    :delete-agent="deleteAgent"
  />
</template>

<script lang="ts" setup>
import { inject, type Ref } from 'vue';

import AgentManager from '~/components/agent/AgentManager.vue';
import useApiClient from '~/compositions/useApiClient';
import type { Agent, Repo } from '~/lib/api/types';

const apiClient = useApiClient();
const repo = inject<Ref<Repo>>('repo');
if (repo === undefined) {
  throw new Error('Unexpected: "repo" should be provided at this place');
}

const loadAgents = (page: number) => apiClient.getRepoAgents(repo.value.id, { page });
const createAgent = (agent: Partial<Agent>) => apiClient.createRepoAgent(repo.value.id, agent);
const updateAgent = (agent: Agent) => apiClient.updateRepoAgent(repo.value.id, agent.id, agent);
const deleteAgent = (agent: Agent) => apiClient.deleteRepoAgent(repo.value.id, agent.id);
</script>
