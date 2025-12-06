<template>
  <form @submit.prevent="$emit('save')">
    <InputField v-slot="{ id }" :label="$t('name')">
      <TextField :id="id" v-model="agent.name" :placeholder="$t('agent_name_placeholder')" required />
    </InputField>

    <InputField :label="$t('agent_no_schedule')">
      <Checkbox
        :model-value="agent.no_schedule || false"
        :label="$t('agent_no_schedule_placeholder')"
        @update:model-value="updateAgent({ no_schedule: $event })"
      />
    </InputField>

    <template v-if="isEditing">
      <InputField v-slot="{ id }" :label="$t('token')">
        <TextField :id="id" v-model="agent.token" :placeholder="$t('token')" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('id')">
        <TextField :id="id" :model-value="agent.id?.toString()" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('backend')" :docs-url="backendDocsUrl">
        <TextField :id="id" v-model="agent.backend" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('platform')">
        <TextField :id="id" v-model="agent.platform" disabled />
      </InputField>

      <InputField
        v-if="agent.custom_labels && Object.keys(agent.custom_labels).length > 0"
        v-slot="{ id }"
        :label="$t('custom_labels')"
      >
        <span class="text-wp-text-alt-100">{{ $t('custom_labels_desc') }}</span>
        <TextField :id="id" :model-value="formatCustomLabels(agent.custom_labels)" disabled />
      </InputField>

      <InputField
        v-slot="{ id }"
        :label="$t('capacity')"
        docs-url="docs/administration/configuration/agent#max_workflows"
      >
        <span class="text-wp-text-alt-100">{{ $t('capacity_desc') }}</span>
        <TextField :id="id" :model-value="agent.capacity?.toString()" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('version')">
        <TextField :id="id" :model-value="agent.version" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('last_contact')">
        <TextField
          :id="id"
          :model-value="agent.last_contact ? date.timeAgo(agent.last_contact * 1000) : $t('never')"
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
        :text="isEditing ? $t('save_agent') : $t('add_agent')"
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
  isEditing?: boolean;
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

const baseDocsUrl = 'https://woodpecker-ci.org/docs/administration/configuration/backends/';

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
