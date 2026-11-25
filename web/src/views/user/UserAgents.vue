<template>
  <AgentManager
    :description="$t('user.settings.agents.desc')"
    :load-agents="loadAgents"
    :create-agent="createAgent"
    :update-agent="updateAgent"
    :delete-agent="deleteAgent"
  >
    <template v-if="org" #headerEnd>
      <RequireOrgAgentsCheckbox
        v-model="org"
        :label="$t('user.settings.agents.require_user_agents.require')"
        :description="$t('user.settings.agents.require_user_agents.desc')"
        :saved-message="$t('user.settings.agents.require_user_agents.saved')"
      />
    </template>
  </AgentManager>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import AgentManager from '~/components/agent/AgentManager.vue';
import RequireOrgAgentsCheckbox from '~/components/agent/RequireOrgAgentsCheckbox.vue';
import useApiClient from '~/compositions/useApiClient';
import useAuthentication from '~/compositions/useAuthentication';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent, Org } from '~/lib/api/types';

const apiClient = useApiClient();
const { user } = useAuthentication();

if (!user) {
  throw new Error('Unexpected: User should be authenticated');
}

const loadAgents = (page: number) => apiClient.getOrgAgents(user.org_id, { page });
const createAgent = (agent: Partial<Agent>) => apiClient.createOrgAgent(user.org_id, agent);
const updateAgent = (agent: Agent) => apiClient.updateOrgAgent(user.org_id, agent.id, agent);
const deleteAgent = (agent: Agent) => apiClient.deleteOrgAgent(user.org_id, agent.id);

const org = ref<Org>();
onMounted(async () => {
  org.value = await apiClient.getOrg(user.org_id);
});

const { t } = useI18n();
useWPTitle(computed(() => [t('admin.settings.agents.agents'), t('user.settings.settings')]));
</script>
