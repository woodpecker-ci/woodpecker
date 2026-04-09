<template>
  <Settings
    :title="$t('registries.registries')"
    :description="$t('org.settings.registries.desc')"
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
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import RegistryEdit from '~/components/registry/RegistryEdit.vue';
import RegistryList from '~/components/registry/RegistryList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Registry } from '~/lib/api/types';
import { deepClone } from '~/lib/utils';

const emptyRegistry: Partial<Registry> = {
  address: '',
  username: '',
  password: '',
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const org = requiredInject('org');
const selectedRegistry = ref<Partial<Registry>>();
const isEditing = computed(() => !!selectedRegistry.value?.id);

async function loadRegistries(page: number, level: 'org' | 'global'): Promise<Registry[] | null> {
  switch (level) {
    case 'org':
      return apiClient.getOrgRegistryList(org.value.id, { page });
    case 'global':
      return apiClient.getGlobalRegistryList({ page });
    default:
      throw new Error(`Unexpected level: ${level}`);
  }
}

const {
  resetPage,
  data: _registries,
  loading,
} = usePagination(loadRegistries, () => !selectedRegistry.value, {
  each: ['org', 'global'],
});
const registries = computed(() => {
  const registriesList: Record<string, Registry & { edit?: boolean; level: 'org' | 'global' }> = {};

  for (const level of ['org', 'global']) {
    for (const registry of _registries.value) {
      if (
        ((level === 'org' && registry.org_id !== 0) || (level === 'global' && registry.org_id === 0)) &&
        !registriesList[registry.address]
      ) {
        registriesList[registry.address] = { ...registry, edit: registry.org_id !== 0, level };
      }
    }
  }

  const levelsOrder = {
    global: 0,
    org: 1,
  };

  return Object.values(registriesList)
    .toSorted((a, b) => a.address.localeCompare(b.address))
    .toSorted((a, b) => levelsOrder[b.level] - levelsOrder[a.level]);
});

const { doSubmit: createRegistry, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedRegistry.value) {
    throw new Error("Unexpected: Can't get registry");
  }

  if (isEditing.value) {
    await apiClient.updateOrgRegistry(org.value.id, selectedRegistry.value);
  } else {
    await apiClient.createOrgRegistry(org.value.id, selectedRegistry.value);
  }
  notifications.notify({
    title: isEditing.value ? i18n.t('registries.saved') : i18n.t('registries.created'),
    type: 'success',
  });
  selectedRegistry.value = undefined;
  await resetPage();
});

const { doSubmit: deleteRegistry, isLoading: isDeleting } = useAsyncAction(async (_registry: Registry) => {
  await apiClient.deleteOrgRegistry(org.value.id, _registry.address);
  notifications.notify({ title: i18n.t('registries.deleted'), type: 'success' });
  await resetPage();
});

function editRegistry(registry: Registry) {
  selectedRegistry.value = deepClone(registry);
}

function showAddRegistry() {
  selectedRegistry.value = deepClone(emptyRegistry);
}

useWPTitle(computed(() => [i18n.t('registries.registries'), org.value.name]));
</script>
