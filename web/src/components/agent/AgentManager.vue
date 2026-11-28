<template>
  <Settings :title="$t('admin.settings.agents.agents')" :description>
    <template #headerActions>
      <Button
        v-if="selectedAgent"
        :text="$t('admin.settings.agents.show')"
        start-icon="back"
        @click="selectedAgent = undefined"
      />
      <Button v-else :text="$t('admin.settings.agents.add')" start-icon="plus" @click="showAddAgent" />
    </template>
    <template v-if="$slots.headerEnd" #headerEnd>
      <slot name="headerEnd" />
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
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import type { Agent } from '~/lib/api/types';
import { deepClone } from '~/lib/utils';

import AgentForm from './AgentForm.vue';
import AgentList from './AgentList.vue';

const props = defineProps<{
  description: string;
  loadAgents: (page: number) => Promise<Agent[] | null>;
  createAgent: (agent: Partial<Agent>) => Promise<Agent>;
  updateAgent: (agent: Agent) => Promise<Agent | void>;
  deleteAgent: (agent: Agent) => Promise<unknown>;
  isAdmin?: boolean;
}>();

const notifications = useNotifications();
const { t } = useI18n();
const route = useRoute();
const router = useRouter();

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
  await resetPage();
});

const { doSubmit: deleteAgent, isLoading: isDeleting } = useAsyncAction(async (_agent: Agent) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('admin.settings.agents.delete_confirm'))) {
    return;
  }

  await props.deleteAgent(_agent);
  notifications.notify({ title: t('admin.settings.agents.deleted'), type: 'success' });
  await resetPage();
});

function editAgent(agent: Agent) {
  selectedAgent.value = deepClone(agent);
}

function showAddAgent() {
  selectedAgent.value = { name: '' };
}

// a link to ?agent=<id>, e.g. from the agent information of a workflow, opens that agent
onMounted(async () => {
  const agentId = Number(route.query.agent);
  if (!agentId) {
    return;
  }
  for (let page = 1; ; page++) {
    const agents = await props.loadAgents(page);
    // the user picked an agent (or add) meanwhile, keep that
    if (selectedAgent.value) {
      return;
    }
    if (!agents?.length) {
      return;
    }
    const agent = agents.find((a) => a.id === agentId);
    if (agent) {
      editAgent(agent);
      return;
    }
  }
});

// keep the shown agent in the url, so a reload opens it again and back on the list it does not
watch(selectedAgent, async (agent) => {
  const agentId = agent?.id?.toString();
  if (route.query.agent === agentId) {
    return;
  }
  const query = { ...route.query };
  if (agentId) {
    query.agent = agentId;
  } else {
    delete query.agent;
  }
  await router.replace({ query });
});
</script>
