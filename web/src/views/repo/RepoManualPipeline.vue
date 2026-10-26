<template>
  <Panel v-if="!loading">
    <form @submit.prevent="triggerManualPipeline">
      <span class="text-wp-text-100 text-xl">{{ $t('repo.manual_pipeline.title') }}</span>

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.message.title')">
        <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.message.desc') }}</span>
        <TextField :id="id" v-model="payload.message" :placeholder="$t('repo.manual_pipeline.message.title')" />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.select_branch')">
        <SelectField :id="id" v-model="payload.branch" :options="branches" required />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.workflow.title')">
        <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.workflow.desc') }}</span>
        <SelectField :id="id" v-model="payload.workflow" :options="workflows" />
        <Warning v-if="workflowsWarning" class="mt-2 text-sm" :text="workflowsWarning" />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('repo.manual_pipeline.variables.title')">
        <span class="text-wp-text-alt-100 mb-2 text-sm">{{ $t('repo.manual_pipeline.variables.desc') }}</span>
        <KeyValueEditor
          :id="id"
          v-model="payload.variables"
          :key-placeholder="$t('repo.manual_pipeline.variables.name')"
          :value-placeholder="$t('repo.manual_pipeline.variables.value')"
          :delete-title="$t('repo.manual_pipeline.variables.delete')"
          @update:is-valid="isVariablesValid = $event"
        />
      </InputField>

      <Button type="submit" :text="$t('repo.manual_pipeline.trigger')" :disabled="!isFormValid" />
    </form>
  </Panel>
  <div v-else class="text-wp-text-100 flex justify-center">
    <Icon name="spinner" />
  </div>
</template>

<script lang="ts" setup>
import { useNotification } from '@kyvg/vue3-notification';
import { computed, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import Warning from '~/components/atomic/Warning.vue';
import InputField from '~/components/form/InputField.vue';
import KeyValueEditor from '~/components/form/KeyValueEditor.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { usePaginate } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { WorkflowInfo } from '~/lib/api/types';

defineProps<{
  open: boolean;
}>();

const emit = defineEmits<{
  (event: 'close'): void;
}>();

const apiClient = useApiClient();
const notifications = useNotification();
const i18n = useI18n();

const repo = requiredInject('repo');
const repoPermissions = requiredInject('repo-permissions');

const router = useRouter();
const branches = ref<{ text: string; value: string }[]>([]);
const workflows = ref<{ text: string; value: string }[]>([]);
const workflowsWarning = ref<string>();
const payload = ref<{ message: string; branch: string; variables: Record<string, string>; workflow: string }>({
  message: '',
  branch: repo.value.default_branch,
  variables: {},
  workflow: '',
});

const isVariablesValid = ref(true);

const isFormValid = computed(() => {
  return payload.value.branch !== '' && isVariablesValid.value;
});

const pipelineOptions = computed(() => ({
  ...payload.value,
  variables: payload.value.variables,
}));

const loading = ref(true);
onMounted(async () => {
  if (!repoPermissions.value.push) {
    notifications.notify({ type: 'error', title: i18n.t('repo.settings.not_allowed') });
    await router.replace({ name: 'home' });
    return;
  }

  const data = await usePaginate((page) => apiClient.getRepoBranches(repo.value.id, { page }));
  branches.value = data.map((e) => ({
    text: e,
    value: e,
  }));
  loading.value = false;
});

async function loadWorkflows(branch: string) {
  const allWorkflows = { text: i18n.t('repo.manual_pipeline.workflow.all'), value: '' };
  payload.value.workflow = '';
  workflows.value = [allWorkflows];
  workflowsWarning.value = undefined;
  if (branch === '') {
    return;
  }

  let data: WorkflowInfo[] | undefined;
  try {
    data = await apiClient.getRepoWorkflows(repo.value.id, branch);
  } catch (error) {
    console.error('Failed to fetch repo workflows', error);
  }

  if (payload.value.branch !== branch) {
    return;
  }
  if (data === undefined) {
    workflowsWarning.value = i18n.t('repo.manual_pipeline.workflow.load_error');
    return;
  }
  workflows.value = [allWorkflows, ...data.map((e) => ({ text: e.name, value: e.name }))];
  if (data.length === 0) {
    workflowsWarning.value = i18n.t('repo.manual_pipeline.no_manual_workflows');
  }
}

watch(
  () => payload.value.branch,
  (branch) => {
    if (repoPermissions.value.push) {
      void loadWorkflows(branch);
    }
  },
  { immediate: true },
);

async function triggerManualPipeline() {
  loading.value = true;
  const pipeline = await apiClient.createPipeline(repo.value.id, pipelineOptions.value);

  emit('close');

  if (typeof pipeline == 'string') {
    // if this is a string (http 204) there is no workflow to run with the 'manual' event

    await router.push({
      name: 'repo',
    });

    notifications.notify({ type: 'warn', title: i18n.t('repo.manual_pipeline.no_manual_workflows') });
  } else {
    await router.push({
      name: 'repo-pipeline',
      params: {
        pipelineId: pipeline.number,
      },
    });
  }

  loading.value = false;
}

useWPTitle(computed(() => [i18n.t('repo.manual_pipeline.trigger'), repo.value.full_name]));
</script>
