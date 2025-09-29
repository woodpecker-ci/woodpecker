<template>
  <Settings
    :title="$t('registries.registries')"
    :description="$t('admin.settings.registries.desc')"
    docs-url="docs/usage/registries"
  >
    <template #headerActions>
      <Button
        v-if="selectedRegistry"
        :text="$t('registries.show')"
        start-icon="back"
        @click="selectedRegistry = undefined"
      />
      <Button v-else :text="$t('registries.add')" start-icon="plus" @click="showAddRegistry" />
    </template>

    <template #headerEnd>
      <Warning class="mt-4 text-sm" :text="$t('admin.settings.registries.warning')" />
    </template>

    <RegistryList
      v-if="!selectedRegistry"
      v-model="registries"
      :is-deleting="isDeleting"
      :loading="loading"
      @edit="editRegistry"
      @delete="deleteRegistry"
    />

    <RegistryEdit
      v-else
      v-model="selectedRegistry"
      :is-saving="isSaving"
      @save="createRegistry"
      @cancel="selectedRegistry = undefined"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Warning from '~/components/atomic/Warning.vue';
import Settings from '~/components/layout/Settings.vue';
import RegistryEdit from '~/components/registry/RegistryEdit.vue';
import RegistryList from '~/components/registry/RegistryList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Registry } from '~/lib/api/types';

const emptyRegistry: Partial<Registry> = {
  address: '',
  username: '',
  password: '',
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const selectedRegistry = ref<Partial<Registry>>();
const isEditingRegistry = computed(() => !!selectedRegistry.value?.id);

async function loadRegistries(page: number): Promise<Registry[] | null> {
  return apiClient.getGlobalRegistryList({ page });
}

const { resetPage, data: registries, loading } = usePagination(loadRegistries, () => !selectedRegistry.value);

const { doSubmit: createRegistry, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedRegistry.value) {
    throw new Error("Unexpected: Can't get registry");
  }

  if (isEditingRegistry.value) {
    await apiClient.updateGlobalRegistry(selectedRegistry.value);
  } else {
    await apiClient.createGlobalRegistry(selectedRegistry.value);
  }
  notifications.notify({
    title: isEditingRegistry.value ? i18n.t('registries.saved') : i18n.t('registries.created'),
    type: 'success',
  });
  selectedRegistry.value = undefined;
  await resetPage();
});

const { doSubmit: deleteRegistry, isLoading: isDeleting } = useAsyncAction(async (_registry: Registry) => {
  await apiClient.deleteGlobalRegistry(_registry.address);
  notifications.notify({ title: i18n.t('registries.deleted'), type: 'success' });
  await resetPage();
});

function editRegistry(registry: Registry) {
  selectedRegistry.value = cloneDeep(registry);
}

function showAddRegistry() {
  selectedRegistry.value = cloneDeep(emptyRegistry);
}

useWPTitle(computed(() => [i18n.t('registries.registries'), i18n.t('admin.settings.settings')]));
</script>
