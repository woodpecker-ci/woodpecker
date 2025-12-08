<template>
  <Settings :title="$t('agents')" :description="$t('agents_server_desc')">
    <template #headerActions>
      <Button :text="$t('show_agents')" start-icon="back" :to="{ name: 'admin-settings-agents' }" />
    </template>

    <AgentForm
      v-model="agent"
      :is-saving="isSaving"
      @save="createAgent"
      @cancel="$router.replace({ name: 'admin-settings-agents' })"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import AgentForm from '~/components/agent/AgentForm.vue';
import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent } from '~/lib/api/types';

const notifications = useNotifications();
const { t } = useI18n();
const apiClient = useApiClient();
const router = useRouter();

const agent = ref<Partial<Agent>>({
  name: '',
  no_schedule: false,
});

const { doSubmit: createAgent, isLoading: isSaving } = useAsyncAction(async () => {
  if (!agent.value) {
    throw new Error("Unexpected: Can't get agent");
  }

  const createdAgent = await apiClient.createAgent(agent.value);

  notifications.notify({
    title: t('agent_created'),
    type: 'success',
  });

  await router.replace({ name: 'admin-settings-agent', params: { agentId: createdAgent.id } });
});

useWPTitle(computed(() => [t('agents'), t('settings'), t('create')]));
</script>
