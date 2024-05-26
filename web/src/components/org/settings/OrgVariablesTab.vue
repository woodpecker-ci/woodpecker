<template>
  <Settings
    :title="$t('variables.variables')"
    :desc="$t('org.settings.variables.desc')"
    docs-url="docs/usage/variables"
  >
    <template #titleActions>
      <Button
        v-if="selectedVariable"
        :text="$t('variables.show')"
        start-icon="back"
        @click="selectedVariable = undefined"
      />
      <Button v-else :text="$t('variables.add')" start-icon="plus" @click="showAddVariable" />
    </template>

    <VariableList
      v-if="!selectedVariable"
      v-model="variables"
      :is-deleting="isDeleting"
      @edit="editVariable"
      @delete="deleteVariable"
    />

    <VariableEdit
      v-else
      v-model="selectedVariable"
      :is-saving="isSaving"
      @save="createVariable"
      @cancel="selectedVariable = undefined"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, inject, ref, type Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import VariableEdit from '~/components/variables/VariableEdit.vue';
import VariableList from '~/components/variables/VariableList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import type { Org, Variable } from '~/lib/api/types';

const emptyVariable: Partial<Variable> = {
  name: '',
  value: '',
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const org = inject<Ref<Org>>('org');
const selectedVariable = ref<Partial<Variable>>();
const isEditingVariable = computed(() => !!selectedVariable.value?.id);

async function loadVariables(page: number): Promise<Variable[] | null> {
  if (!org?.value) {
    throw new Error("Unexpected: Can't load org");
  }

  return apiClient.getOrgVariableList(org.value.id, { page });
}

const { resetPage, data: variables } = usePagination(loadVariables, () => !selectedVariable.value);

const { doSubmit: createVariable, isLoading: isSaving } = useAsyncAction(async () => {
  if (!org?.value) {
    throw new Error("Unexpected: Can't load org");
  }

  if (!selectedVariable.value) {
    throw new Error("Unexpected: Can't get variable");
  }

  if (isEditingVariable.value) {
    await apiClient.updateOrgVariable(org.value.id, selectedVariable.value);
  } else {
    await apiClient.createOrgVariable(org.value.id, selectedVariable.value);
  }
  notifications.notify({
    title: isEditingVariable.value ? i18n.t('variables.saved') : i18n.t('variables.created'),
    type: 'success',
  });
  selectedVariable.value = undefined;
  resetPage();
});

const { doSubmit: deleteVariable, isLoading: isDeleting } = useAsyncAction(async (_variable: Variable) => {
  if (!org?.value) {
    throw new Error("Unexpected: Can't load org");
  }

  await apiClient.deleteOrgVariable(org.value.id, _variable.name);
  notifications.notify({ title: i18n.t('variables.deleted'), type: 'success' });
  resetPage();
});

function editVariable(variable: Variable) {
  selectedVariable.value = cloneDeep(variable);
}

function showAddVariable() {
  selectedVariable.value = cloneDeep(emptyVariable);
}
</script>
