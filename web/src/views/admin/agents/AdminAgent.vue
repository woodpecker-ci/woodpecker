<template>
  <Settings :title="$t('agents')" :description="$t('agents_server_desc')">
    <template #headerActions>
      <Button :text="$t('show_agents')" start-icon="back" :to="{ name: 'admin-settings-agents' }" />
    </template>

    <AgentForm
      v-if="agent"
      v-model="agent"
      is-editing
      :is-saving="isSaving"
      @save="saveAgent"
      @cancel="$router.replace({ name: 'admin-settings-agents' })"
    />
    <div v-else class="flex justify-center">
      <Icon name="spinner" class="animate-spin" />
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import AgentForm from '~/components/agent/AgentForm.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction, useAsyncData } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent } from '~/lib/api/types';

const notifications = useNotifications();
const { t } = useI18n();
const apiClient = useApiClient();
const route = useRoute();

const agentId = computed(() => Number.parseInt(route.params.agentId.toString(), 10));

const agent = ref<Agent | null>(null);

const { data: dbAgent, refetch: reloadAgent } = useAsyncData(computed(() => () => apiClient.getAgent(agentId.value)));

watch(
  dbAgent,
  (newAgent) => {
    agent.value = newAgent;
  },
  { immediate: true },
);

const { doSubmit: saveAgent, isLoading: isSaving } = useAsyncAction(async () => {
  if (!agent.value) {
    throw new Error("Unexpected: Can't get agent");
  }

  await apiClient.updateAgent(agent.value);

  notifications.notify({
    title: t('agent_saved'),
    type: 'success',
  });

  await reloadAgent();
});

useWPTitle(computed(() => [t('agents'), t('settings'), agent.value?.name ?? '']));
</script>
