<template>
  <form @submit.prevent="$emit('save')">
    <InputField v-slot="{ id }" :label="$t('admin.settings.agents.name.name')">
      <TextField :id="id" v-model="agent.name" :placeholder="$t('admin.settings.agents.name.placeholder')" required />
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
        :model-value="agent.no_schedule || false"
        :label="$t('admin.settings.agents.no_schedule.placeholder')"
        @update:model-value="updateAgent({ no_schedule: $event })"
      />
    </InputField>

    <template v-if="isEditingAgent">
      <InputField v-slot="{ id }" :label="$t('admin.settings.agents.token')">
        <TextField :id="id" v-model="agent.token" :placeholder="$t('admin.settings.agents.token')" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('admin.settings.agents.id')">
        <TextField :id="id" :model-value="agent.id?.toString()" disabled />
      </InputField>

      <InputField
        v-slot="{ id }"
        :label="$t('admin.settings.agents.backend.backend')"
        docs-url="docs/next/administration/backends/docker"
      >
        <TextField :id="id" v-model="agent.backend" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('admin.settings.agents.platform.platform')">
        <TextField :id="id" v-model="agent.platform" disabled />
      </InputField>

      <InputField
        v-slot="{ id }"
        :label="$t('admin.settings.agents.capacity.capacity')"
        docs-url="docs/next/administration/agent-config#woodpecker_max_workflows"
      >
        <span class="text-wp-text-alt-100">{{ $t('admin.settings.agents.capacity.desc') }}</span>
        <TextField :id="id" :model-value="agent.capacity?.toString()" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('admin.settings.agents.version')">
        <TextField :id="id" :model-value="agent.version" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('admin.settings.agents.last_contact')">
        <TextField
          :id="id"
          :model-value="
            agent.last_contact ? date.timeAgo(agent.last_contact * 1000) : $t('admin.settings.agents.never')
          "
          disabled
        />
      </InputField>
    </template>

    <div class="flex gap-2">
      <Button type="button" color="gray" :text="$t('cancel')" @click="$emit('cancel')" />
      <Button
        :is-loading="isSaving"
        type="submit"
        color="green"
        :text="isEditingAgent ? $t('admin.settings.agents.save') : $t('admin.settings.agents.add')"
      />
    </div>
  </form>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import { useDate } from '~/compositions/useDate';
import type { Agent } from '~/lib/api/types';

const props = defineProps<{
  modelValue: Partial<Agent>;
  isEditingAgent: boolean;
  isSaving: boolean;
}>();

const emit = defineEmits<{
  (e: 'update:modelValue', value: Partial<Agent>): void;
  (e: 'save'): void;
  (e: 'cancel'): void;
}>();

const date = useDate();

const agent = computed({
  get: () => props.modelValue,
  set: (value) => emit('update:modelValue', value),
});
const filters = computed(() => Object.entries(agent.value.filters || {}).map(([key, value]) => ({ key, value })));

const newFilterKey = ref('');
const newFilterValue = ref('');

function updateAgent(newValues: Partial<Agent>) {
  emit('update:modelValue', { ...agent.value, ...newValues });
}

function deleteFilter(index: number) {
  const newFilters = [...filters.value];
  newFilters.splice(index, 1);
  updateFilters(newFilters);
}

function addFilter() {
  if (newFilterKey.value && newFilterValue.value) {
    const newFilters = [...filters.value, { key: newFilterKey.value, value: newFilterValue.value }];
    updateFilters(newFilters);
    newFilterKey.value = '';
    newFilterValue.value = '';
  }
}

function updateFilters(newFilters: Array<{ key: string; value: string }>) {
  const filtersObject = Object.fromEntries(newFilters.map((filter) => [filter.key, filter.value]));
  updateAgent({ filters: filtersObject });
}
</script>
