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
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import AgentManager from '~/components/agent/AgentManager.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent } from '~/lib/api/types';

const apiClient = useApiClient();
const org = requiredInject('org');
const org = requiredInject('org');

const loadAgents = (page: number) => apiClient.getOrgAgents(org.value.id, { page });
const createAgent = (agent: Partial<Agent>) => apiClient.createOrgAgent(org.value.id, agent);
const updateAgent = (agent: Agent) => apiClient.updateOrgAgent(org.value.id, agent.id, agent);
const deleteAgent = (agent: Agent) => apiClient.deleteOrgAgent(org.value.id, agent.id);

const { t } = useI18n();
useWPTitle(computed(() => [t('admin.settings.agents.agents'), org.value.name]));
</script>
