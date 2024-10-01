<template>
  <Settings :title="$t('admin.settings.agents.agents')" :desc="desc">
    <template #titleActions>
      <Button
        v-if="selectedAgent"
        :text="$t('admin.settings.agents.show')"
        start-icon="back"
        @click="selectedAgent = undefined"
      />
      <Button v-else :text="$t('admin.settings.agents.add')" start-icon="plus" @click="showAddAgent" />
    </template>

    <AgentList
      v-if="!selectedAgent"
      :loading="loading"
      :agents="agents"
      :is-deleting="isDeleting"
      :is-admin="isAdmin"
      @edit="editAgent"
      @delete="deleteAgent"
    />
    <AgentForm
      v-else
      v-model="selectedAgent"
      :is-editing-agent="isEditingAgent"
      :is-saving="isSaving"
      @save="saveAgent"
      @cancel="selectedAgent = undefined"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import type { Agent } from '~/lib/api/types';

import AgentForm from './AgentForm.vue';
import AgentList from './AgentList.vue';

const props = defineProps<{
  desc: string;
  loadAgents: (page: number) => Promise<Agent[] | null>;
  createAgent: (agent: Partial<Agent>) => Promise<Agent>;
  updateAgent: (agent: Agent) => Promise<Agent | void>;
  deleteAgent: (agent: Agent) => Promise<unknown>;
  isAdmin?: boolean;
}>();

const notifications = useNotifications();
const { t } = useI18n();

const selectedAgent = ref<Partial<Agent>>();
const isEditingAgent = computed(() => !!selectedAgent.value?.id);

const { resetPage, data: agents, loading } = usePagination(props.loadAgents);

const { doSubmit: saveAgent, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedAgent.value) {
    throw new Error("Unexpected: Can't get agent");
  }

  if (isEditingAgent.value) {
    await props.updateAgent(selectedAgent.value as Agent);
    selectedAgent.value = undefined;
  } else {
    selectedAgent.value = await props.createAgent(selectedAgent.value);
  }
  notifications.notify({
    title: isEditingAgent.value ? t('admin.settings.agents.saved') : t('admin.settings.agents.created'),
    type: 'success',
  });
  resetPage();
});

const { doSubmit: deleteAgent, isLoading: isDeleting } = useAsyncAction(async (_agent: Agent) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('admin.settings.agents.delete_confirm'))) {
    return;
  }

  await props.deleteAgent(_agent);
  notifications.notify({ title: t('admin.settings.agents.deleted'), type: 'success' });
  resetPage();
});

function editAgent(agent: Agent) {
  selectedAgent.value = cloneDeep(agent);
}

function showAddAgent() {
  selectedAgent.value = { name: '' };
}
</script>
