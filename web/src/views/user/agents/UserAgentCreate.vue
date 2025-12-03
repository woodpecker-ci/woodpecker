<template>
  <Settings :title="$t('admin.settings.agents.agents')" :description="$t('admin.settings.agents.desc')">
    <template #headerActions>
      <Button :text="$t('admin.settings.agents.show')" start-icon="back" :to="{ name: 'user-settings-agents' }" />
    </template>

    <AgentForm
      v-model="agent"
      :is-saving="isSaving"
      @save="createAgent"
      @cancel="$router.replace({ name: 'user-settings-agents' })"
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
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Agent } from '~/lib/api/types';

const notifications = useNotifications();
const { t } = useI18n();
const apiClient = useApiClient();
const router = useRouter();
const { user } = useAuthentication();

if (!user) {
  throw new Error('Unexpected: User should be authenticated');
}

const agent = ref<Partial<Agent>>({
  name: '',
  no_schedule: false,
});

const { doSubmit: createAgent, isLoading: isSaving } = useAsyncAction(async () => {
  if (!agent.value) {
    throw new Error("Unexpected: Can't get agent");
  }

  const createdAgent = await apiClient.createOrgAgent(user.org_id, agent.value);

  notifications.notify({
    title: t('admin.settings.agents.created'),
    type: 'success',
  });

  await router.push({ name: 'user-settings-agent', params: { agentId: createdAgent.id } });
});

useWPTitle(computed(() => [t('admin.settings.agents.agents'), t('admin.settings.settings'), t('create')]));
</script>
