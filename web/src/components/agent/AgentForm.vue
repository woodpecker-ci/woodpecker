<template>
  <form @submit.prevent="$emit('save')">
    <InputField v-slot="{ id }" :label="$t('admin.settings.agents.name.name')">
      <TextField :id="id" v-model="agent.name" :placeholder="$t('admin.settings.agents.name.placeholder')" required />
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

      <InputField v-slot="{ id }" :label="$t('admin.settings.agents.backend.backend')" :docs-url="backendDocsUrl">
        <TextField :id="id" v-model="agent.backend" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('admin.settings.agents.platform.platform')">
        <TextField :id="id" v-model="agent.platform" disabled />
      </InputField>

      <InputField
        v-if="agent.custom_labels && Object.keys(agent.custom_labels).length > 0"
        v-slot="{ id }"
        :label="$t('admin.settings.agents.custom_labels.custom_labels')"
      >
        <span class="text-wp-text-alt-100">{{ $t('admin.settings.agents.custom_labels.desc') }}</span>
        <TextField :id="id" :model-value="formatCustomLabels(agent.custom_labels)" disabled />
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
import { computed } from 'vue';

import Button from '~/components/atomic/Button.vue';
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

const baseDocsUrl = 'https://woodpecker-ci.org/docs/next/administration/backends/';

const backendDocsUrl = computed(() => {
  let backendUrlSuffix = agent.value.backend?.toLowerCase();
  if (backendUrlSuffix === 'custom') {
    backendUrlSuffix = 'custom-backends';
  }
  return `${baseDocsUrl}${backendUrlSuffix === '' ? 'docker' : backendUrlSuffix}`;
});

function updateAgent(newValues: Partial<Agent>) {
  emit('update:modelValue', { ...agent.value, ...newValues });
}

function formatCustomLabels(labels: Record<string, string>): string {
  return Object.entries(labels)
    .map(([key, value]) => `${key}=${value}`)
    .join(', ');
}
</script>
