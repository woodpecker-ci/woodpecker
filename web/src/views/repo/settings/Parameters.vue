<template>
  <Settings :title="$t('parameters.parameters')" :description="$t('parameters.desc')" docs-url="docs/usage/parameters">
    <template #headerActions>
      <Button
        v-if="selectedParameter"
        :text="$t('parameters.show')"
        start-icon="back"
        @click="selectedParameter = undefined"
      />
      <Button
        v-else
        :text="$t('parameters.add')"
        start-icon="plus"
        @click="showAddParameter()"
      />
    </template>

    <ParameterList
      v-if="!selectedParameter"
      :parameters="parameters"
      :is-deleting="isDeleting"
      @edit="editParameter"
      @delete="deleteParameter"
    />

    <ParameterEdit
      v-else
      v-model="selectedParameter"
      :existing-parameter="isEditingParameter"
      :is-saving="isSaving"
      @save="createParameter"
      @cancel="selectedParameter = undefined"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, inject, ref } from 'vue';
import type { Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import ParameterEdit from '~/components/parameters/ParameterEdit.vue';
import ParameterList from '~/components/parameters/ParameterList.vue';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import type { Parameter, Repo } from '~/lib/api/types';
import { ParameterType } from '~/lib/api/types';
import useApiClient from '~/compositions/useApiClient';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();
const repo = inject('repo') as Ref<Repo>;

const parameters = ref<Parameter[]>([]);
const selectedParameter = ref<Partial<Parameter>>();
const emptyParameter: Partial<Parameter> = {
  repo_id: repo.value?.id,
  name: '',
  branch: '*',
  type: ParameterType.String,
  description: '',
  default_value: '',
  trim_string: true,
};

const isEditingParameter = computed(() => selectedParameter.value?.id !== undefined);

async function resetPage() {
  parameters.value = (await apiClient.getParameters(repo.value)) || [];
}

function showAddParameter() {
  selectedParameter.value = cloneDeep(emptyParameter);
}

const { doSubmit: createParameter, isLoading: isSaving } = useAsyncAction(async () => {
  if (!repo?.value || !selectedParameter.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  if (isEditingParameter.value) {
    await apiClient.updateParameter(repo.value, selectedParameter.value);
  } else {
    await apiClient.createParameter(repo.value, selectedParameter.value);
  }
  notifications.notify({
    title: isEditingParameter.value ? i18n.t('parameters.saved') : i18n.t('parameters.created'),
    type: 'success',
  });
  selectedParameter.value = undefined;
  await resetPage();
});

const { doSubmit: deleteParameter, isLoading: isDeleting } = useAsyncAction(async (_parameter: Parameter) => {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  await apiClient.deleteParameter(repo.value, _parameter.id);
  notifications.notify({ title: i18n.t('parameters.deleted'), type: 'success' });
  await resetPage();
});

function editParameter(parameter: Parameter) {
  selectedParameter.value = cloneDeep(parameter);
}

resetPage();
</script>
