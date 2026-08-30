<template>
  <InputField v-slot="{ id }" :label="$t('repo.workflow_select.title')">
    <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.workflow_select.desc') }}</span>

    <div v-if="loading" class="text-wp-text-alt-100 flex items-center gap-2 text-sm">
      <Icon name="spinner" />
      <span>{{ $t('repo.workflow_select.loading') }}</span>
    </div>

    <span v-else-if="error" class="text-wp-error-100 text-sm">{{ error }}</span>

    <span v-else-if="options.length === 0" class="text-wp-text-alt-100 text-sm">
      {{ $t('repo.workflow_select.none') }}
    </span>

    <CheckboxesField v-else :id="id" v-model="innerValue" :options="options" />
  </InputField>
</template>

<script lang="ts" setup>
import { computed, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import type { CheckboxOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import useApiClient from '~/compositions/useApiClient';

const props = defineProps<{
  modelValue?: string[];
  repoId: number;
  // branch to read the workflow list from, defaults to the repo default branch
  branch?: string;
}>();

const emit = defineEmits<{
  (event: 'update:modelValue', value: string[]): void;
}>();

const apiClient = useApiClient();
const i18n = useI18n();

const modelValue = toRef(props, 'modelValue');
const innerValue = computed({
  get: () => modelValue.value ?? [],
  set: (value) => emit('update:modelValue', value),
});

const options = ref<CheckboxOption[]>([]);
const loading = ref(false);
const error = ref('');

async function loadWorkflows(branch?: string) {
  loading.value = true;
  error.value = '';
  try {
    const workflows = await apiClient.getRepoWorkflows(props.repoId, branch);
    options.value = workflows.map((workflow) => ({ text: workflow, value: workflow }));

    // a workflow that no longer exists on this branch cannot stay selected
    const available = new Set(workflows);
    const stillValid = innerValue.value.filter((workflow) => available.has(workflow));
    if (stillValid.length !== innerValue.value.length) {
      innerValue.value = stillValid;
    }
  } catch {
    options.value = [];
    error.value = i18n.t('repo.workflow_select.error');
  } finally {
    loading.value = false;
  }
}

watch(() => props.branch, loadWorkflows, { immediate: true });
</script>
