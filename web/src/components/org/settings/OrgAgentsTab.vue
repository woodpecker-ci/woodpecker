<template>
  <Settings :title="$t('admin.settings.agents.agents')" :desc="$t('admin.settings.agents.desc')">
    <template #titleActions>
      <Button
        v-if="selectedAgent"
        :text="$t('admin.settings.agents.show')"
        start-icon="back"
        @click="selectedAgent = undefined"
      />
      <Button v-else :text="$t('admin.settings.agents.add')" start-icon="plus" @click="showAddAgent" />
    </template>

    <div v-if="!selectedAgent" class="space-y-4 text-wp-text-100">
      <ListItem
        v-for="agent in agents"
        :key="agent.id"
        class="items-center !bg-wp-background-200 !dark:bg-wp-background-100"
      >
        <span>{{ agent.name || `Agent ${agent.id}` }}</span>
        <span class="ml-auto">
          <span class="hidden md:inline-block space-x-2">
            <Badge v-if="agent.platform" :label="$t('admin.settings.agents.platform.badge')" :value="agent.platform" />
            <Badge v-if="agent.backend" :label="$t('admin.settings.agents.backend.badge')" :value="agent.backend" />
            <Badge v-if="agent.capacity" :label="$t('admin.settings.agents.capacity.badge')" :value="agent.capacity" />
          </span>
          <span class="ml-2">{{
            agent.last_contact ? date.timeAgo(agent.last_contact * 1000) : $t('admin.settings.agents.never')
          }}</span>
        </span>
        <IconButton
          icon="edit"
          :title="$t('admin.settings.agents.edit_agent')"
          class="ml-2 w-8 h-8"
          @click="editAgent(agent)"
        />
        <IconButton
          icon="trash"
          :title="$t('admin.settings.agents.delete_agent')"
          class="ml-2 w-8 h-8 hover:text-wp-control-error-100"
          :is-loading="isDeleting"
          @click="deleteAgent(agent)"
        />
      </ListItem>

      <div v-if="agents?.length === 0" class="ml-2">{{ $t('admin.settings.agents.none') }}</div>
    </div>
    <div v-else>
      <form @submit.prevent="saveAgent">
        <InputField v-slot="{ id }" :label="$t('admin.settings.agents.name.name')">
          <TextField
            :id="id"
            v-model="selectedAgent.name"
            :placeholder="$t('admin.settings.agents.name.placeholder')"
            required
          />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('admin.settings.agents.filters.name')">
          <span class="text-sm text-wp-text-alt-100 mb-2">{{ $t('admin.settings.agents.filters.desc') }}</span>
          <div class="flex flex-col gap-2">
            <div v-for="(filter, index) in filters" :key="index" class="flex gap-4">
              <TextField
                :id="`${id}-key-${index}`"
                v-model="filter.key"
                :placeholder="$t('admin.settings.agents.filters.key')"
              />
              <TextField
                :id="`${id}-value-${index}`"
                v-model="filter.value"
                :placeholder="$t('admin.settings.agents.filters.value')"
              />
              <div class="w-10 flex-shrink-0">
                <Button
                  type="button"
                  color="red"
                  class="ml-auto"
                  :title="$t('admin.settings.agents.filters.delete')"
                  @click="deleteFilter(index)"
                >
                  <Icon name="remove" />
                </Button>
              </div>
            </div>
            <div class="flex gap-4">
              <TextField
                :id="`${id}-new-key`"
                v-model="newFilterKey"
                :placeholder="$t('admin.settings.agents.filters.key')"
              />
              <TextField
                :id="`${id}-new-value`"
                v-model="newFilterValue"
                :placeholder="$t('admin.settings.agents.filters.value')"
              />
              <div class="w-10 flex-shrink-0">
                <Button
                  type="button"
                  color="green"
                  class="ml-auto"
                  :title="$t('admin.settings.agents.filters.add')"
                  @click="addFilter"
                >
                  <Icon name="plus" />
                </Button>
              </div>
            </div>
          </div>
        </InputField>

        <InputField :label="$t('admin.settings.agents.no_schedule.name')">
          <Checkbox
            :model-value="selectedAgent.no_schedule || false"
            :label="$t('admin.settings.agents.no_schedule.placeholder')"
            @update:model-value="selectedAgent!.no_schedule = $event"
          />
        </InputField>

        <template v-if="isEditingAgent">
          <InputField v-slot="{ id }" :label="$t('admin.settings.agents.token')">
            <TextField
              :id="id"
              v-model="selectedAgent.token"
              :placeholder="$t('admin.settings.agents.token')"
              disabled
            />
          </InputField>

          <InputField v-slot="{ id }" :label="$t('admin.settings.agents.id')">
            <TextField :id="id" :model-value="selectedAgent.id?.toString()" disabled />
          </InputField>

          <InputField
            v-slot="{ id }"
            :label="$t('admin.settings.agents.backend.backend')"
            docs-url="docs/next/administration/backends/docker"
          >
            <TextField :id="id" v-model="selectedAgent.backend" disabled />
          </InputField>

          <InputField v-slot="{ id }" :label="$t('admin.settings.agents.platform.platform')">
            <TextField :id="id" v-model="selectedAgent.platform" disabled />
          </InputField>

          <InputField
            v-slot="{ id }"
            :label="$t('admin.settings.agents.capacity.capacity')"
            docs-url="docs/next/administration/agent-config#woodpecker_max_workflows"
          >
            <span class="text-wp-text-alt-100">{{ $t('admin.settings.agents.capacity.desc') }}</span>
            <TextField :id="id" :model-value="selectedAgent.capacity?.toString()" disabled />
          </InputField>

          <InputField v-slot="{ id }" :label="$t('admin.settings.agents.version')">
            <TextField :id="id" :model-value="selectedAgent.version" disabled />
          </InputField>

          <InputField v-slot="{ id }" :label="$t('admin.settings.agents.last_contact')">
            <TextField
              :id="id"
              :model-value="
                selectedAgent.last_contact
                  ? date.timeAgo(selectedAgent.last_contact * 1000)
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
  </Settings>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, inject, ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useDate } from '~/compositions/useDate';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import type { Agent, Org } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const date = useDate();
const { t } = useI18n();

const org = inject<Ref<Org>>('org');
if (org === undefined) {
  throw new Error('Unexpected: "org" should be provided at this place');
}

const selectedAgent = ref<Partial<Agent>>();
const isEditingAgent = computed(() => !!selectedAgent.value?.id);
const filters = ref<Array<{ key: string; value: string }>>([]);
const newFilterKey = ref('');
const newFilterValue = ref('');

async function loadAgents(page: number): Promise<Agent[] | null> {
  return apiClient.getOrgAgents(org!.value.id, { page });
}

const { resetPage, data: agents } = usePagination(loadAgents, () => !selectedAgent.value);

const { doSubmit: saveAgent, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedAgent.value) {
    throw new Error("Unexpected: Can't get agent");
  }

  selectedAgent.value.filters = Object.fromEntries(filters.value.map((filter) => [filter.key, filter.value]));

  if (isEditingAgent.value) {
    await apiClient.updateOrgAgent(org.value.id, selectedAgent.value.id!, selectedAgent.value);
    selectedAgent.value = undefined;
    filters.value = [];
  } else {
    selectedAgent.value = await apiClient.createOrgAgent(org.value.id, selectedAgent.value);
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

  await apiClient.deleteOrgAgent(org.value.id, _agent.id);
  notifications.notify({ title: t('admin.settings.agents.deleted'), type: 'success' });
  resetPage();
});

function editAgent(agent: Agent) {
  selectedAgent.value = cloneDeep(agent);
  filters.value = Object.entries(agent.filters || {}).map(([key, value]) => ({ key, value }));
}

function showAddAgent() {
  selectedAgent.value = cloneDeep({ name: '', filters: {} });
  filters.value = [];
}

function deleteFilter(index: number) {
  filters.value.splice(index, 1);
}

function addFilter() {
  if (newFilterKey.value && newFilterValue.value) {
    filters.value.push({ key: newFilterKey.value, value: newFilterValue.value });
    newFilterKey.value = '';
    newFilterValue.value = '';
  }
}
</script>
