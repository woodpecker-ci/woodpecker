<template>
  <Settings :title="$t('admin.settings.agents.agents')" :description="$t('admin.settings.agents.desc')">
    <template #headerActions>
      <Button :text="$t('admin.settings.agents.show')" start-icon="back" @click="$router.back()" />
    </template>

    <AgentForm
      v-if="agent"
      v-model="agent"
      is-editing-agent
      :is-saving="isSaving"
      @save="saveAgent"
      @cancel="$router.back()"
    />
    <!-- TODO: show loading spinner -->
  </Settings>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute } from 'vue-router';

import AgentForm from '~/components/agent/AgentForm.vue';
import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction, useAsyncData } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { useWPTitle } from '~/compositions/useWPTitle';

const notifications = useNotifications();
const { t } = useI18n();
const apiClient = useApiClient();
const route = useRoute();

const agentId = computed(() => Number.parseInt(route.params.agentId.toString(), 10));

const { data: agent, refetch: reloadAgent } = await useAsyncData(apiClient.getAgent, [agentId]);

const { doSubmit: saveAgent, isLoading: isSaving } = useAsyncAction(async () => {
  if (!agent.value) {
    throw new Error("Unexpected: Can't get agent");
  }

  await apiClient.updateAgent(agent.value);

  notifications.notify({
    title: t('admin.settings.agents.saved'),
    type: 'success',
  });

  await reloadAgent();
});

useWPTitle(computed(() => [t('admin.settings.agents.agents'), t('admin.settings.settings'), agent.value?.name ?? '']));
</script>
