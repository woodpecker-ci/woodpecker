<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-color">{{ $t('admin.settings.agents.agents') }}</h1>
        <p class="text-sm text-color-alt">{{ $t('admin.settings.agents.desc') }}</p>
      </div>
      <Button
        v-if="selectedAgent"
        class="ml-auto"
        :text="$t('admin.settings.agents.show')"
        start-icon="back"
        @click="selectedAgent = undefined"
      />
      <template v-else>
        <Button class="ml-auto" :text="$t('admin.settings.agents.add')" start-icon="plus" @click="showAddAgent" />
        <Button class="ml-2" start-icon="refresh" @click="loadAgents" />
      </template>
    </div>

    <div v-if="!selectedAgent" class="space-y-4 text-color">
      <ListItem v-for="agent in agents" :key="agent.id" class="items-center">
        <span>{{ agent.name || `Agent ${agent.id}` }}</span>
        <span class="ml-auto">
          <span class="hidden md:inline-block space-x-2">
            <Badge :label="$t('admin.settings.agents.platform.badge')" :value="agent.platform" />
            <Badge :label="$t('admin.settings.agents.backend.badge')" :value="agent.backend" />
            <Badge :label="$t('admin.settings.agents.capacity.badge')" :value="agent.capacity" />
          </span>
          <span class="ml-2">{{ agent.last_contact ? timeAgo.format(agent.last_contact * 1000) : 'never' }}</span>
        </span>
        <IconButton icon="edit" class="ml-2 w-8 h-8" @click="editAgent(agent)" />
        <IconButton
          icon="trash"
          class="ml-2 w-8 h-8 hover:text-red-400 hover:dark:text-red-500"
          :is-loading="isDeleting"
          @click="deleteAgent(agent)"
        />
      </ListItem>

      <div v-if="agents?.length === 0" class="ml-2">{{ $t('admin.settings.agents.none') }}</div>
    </div>
    <div v-else>
      <form @submit.prevent="saveAgent">
        <InputField :label="$t('admin.settings.agents.name.name')">
          <TextField
            v-model="selectedAgent.name"
            :placeholder="$t('admin.settings.agents.name.placeholder')"
            required
          />
        </InputField>

        <InputField :label="$t('admin.settings.agents.no_schedule.name')">
          <Checkbox
            :model-value="selectedAgent.no_schedule || false"
            :label="$t('admin.settings.agents.no_schedule.placeholder')"
            @update:model-value="selectedAgent!.no_schedule = $event"
          />
        </InputField>

        <template v-if="isEditingAgent">
          <InputField :label="$t('admin.settings.agents.token')">
            <TextField v-model="selectedAgent.token" :placeholder="$t('admin.settings.agents.token')" disabled />
          </InputField>

          <InputField
            :label="$t('admin.settings.agents.backend.backend')"
            docs-url="docs/next/administration/backends/docker"
          >
            <TextField v-model="selectedAgent.backend" disabled />
          </InputField>

          <InputField :label="$t('admin.settings.agents.platform.platform')">
            <TextField v-model="selectedAgent.platform" disabled />
          </InputField>

          <InputField
            :label="$t('admin.settings.agents.capacity.capacity')"
            docs-url="docs/next/administration/agent-config#woodpecker_max_procs"
          >
            <span class="text-color-alt">{{ $t('admin.settings.agents.capacity.desc') }}</span>
            <TextField :model-value="selectedAgent.capacity?.toString()" disabled />
          </InputField>

          <InputField :label="$t('admin.settings.agents.version')">
            <TextField :model-value="selectedAgent.version" disabled />
          </InputField>

          <InputField :label="$t('admin.settings.agents.last_contact')">
            <TextField
              :model-value="
                selectedAgent.last_contact
                  ? timeAgo.format(selectedAgent.last_contact * 1000)
                  : $t('admin.settings.agents.never')
              "
              disabled
            />
          </InputField>
        </template>

        <div class="flex gap-2">
          <Button type="button" color="gray" :text="$t('cancel')" @click="selectedAgent = undefined" />
          <Button
            :is-loading="isSaving"
            type="submit"
            color="green"
            :text="isEditingAgent ? $t('admin.settings.agents.save') : $t('admin.settings.agents.add')"
          />
        </div>
      </form>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { Agent } from '~/lib/api/types';
import timeAgo from '~/utils/timeAgo';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const agents = ref<Agent[]>([]);
const selectedAgent = ref<Partial<Agent>>();
const isEditingAgent = computed(() => !!selectedAgent.value?.id);

async function loadAgents() {
  agents.value = await apiClient.getAgents();
}

const { doSubmit: saveAgent, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedAgent.value) {
    throw new Error("Unexpected: Can't get agent");
  }

  if (isEditingAgent.value) {
    await apiClient.updateAgent(selectedAgent.value);
    selectedAgent.value = undefined;
  } else {
    selectedAgent.value = await apiClient.createAgent(selectedAgent.value);
  }
  notifications.notify({
    title: i18n.t(isEditingAgent.value ? 'admin.settings.agents.saved' : 'admin.settings.agents.created'),
    type: 'success',
  });
  await loadAgents();
});

const { doSubmit: deleteAgent, isLoading: isDeleting } = useAsyncAction(async (_agent: Agent) => {
  // eslint-disable-next-line no-restricted-globals, no-alert
  if (!confirm(i18n.t('admin.settings.agents.delete_confirm'))) {
    return;
  }

  await apiClient.deleteAgent(_agent);
  notifications.notify({ title: i18n.t('admin.settings.agents.deleted'), type: 'success' });
  await loadAgents();
});

function editAgent(agent: Agent) {
  selectedAgent.value = cloneDeep(agent);
}

function showAddAgent() {
  selectedAgent.value = cloneDeep({ name: '' });
}

onMounted(async () => {
  await loadAgents();
});
</script>
