<template>
  <Settings :title="$t('secrets.secrets')" :description="$t('secrets.desc')" docs-url="docs/usage/secrets">
    <template #headerActions>
      <Button v-if="selectedSecret" :text="$t('secrets.show')" start-icon="back" @click="selectedSecret = undefined" />
      <Button v-else :text="$t('secrets.add')" start-icon="plus" @click="showAddSecret" />
    </template>

    <SecretList
      v-if="!selectedSecret"
      :model-value="secrets"
      :is-deleting="isDeleting"
      @edit="editSecret"
      @delete="deleteSecret"
    />

    <SecretEdit
      v-else
      v-model="selectedSecret"
      :is-saving="isSaving"
      @save="createSecret"
      @cancel="selectedSecret = undefined"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import SecretEdit from '~/components/secrets/SecretEdit.vue';
import SecretList from '~/components/secrets/SecretList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { WebhookEvents } from '~/lib/api/types';
import type { Secret } from '~/lib/api/types';

const emptySecret: Partial<Secret> = {
  name: '',
  value: '',
  images: [],
  events: [WebhookEvents.Push],
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const repo = requiredInject('repo');
const selectedSecret = ref<Partial<Secret>>();
const isEditingSecret = computed(() => !!selectedSecret.value?.id);

async function loadSecrets(page: number, level: 'repo' | 'org' | 'global'): Promise<Secret[] | null> {
  switch (level) {
    case 'repo':
      return apiClient.getSecretList(repo.value.id, { page });
    case 'org':
      return apiClient.getOrgSecretList(repo.value.org_id, { page });
    case 'global':
      return apiClient.getGlobalSecretList({ page });
    default:
      throw new Error(`Unexpected level: ${level}`);
  }
}

const { resetPage, data: _secrets } = usePagination(loadSecrets, () => !selectedSecret.value, {
  each: ['repo', 'org', 'global'],
});
const secrets = computed(() => {
  const secretsList: Record<string, Secret & { edit?: boolean; level: 'repo' | 'org' | 'global' }> = {};

  for (const level of ['repo', 'org', 'global']) {
    for (const secret of _secrets.value) {
      if (
        ((level === 'repo' && secret.repo_id !== 0 && secret.org_id === 0) ||
          (level === 'org' && secret.repo_id === 0 && secret.org_id !== 0) ||
          (level === 'global' && secret.repo_id === 0 && secret.org_id === 0)) &&
        !secretsList[secret.name]
      ) {
        secretsList[secret.name] = { ...secret, edit: secret.repo_id !== 0, level };
      }
    }
  }

  const levelsOrder = {
    global: 0,
    org: 1,
    repo: 2,
  };

  return Object.values(secretsList)
    .toSorted((a, b) => a.name.localeCompare(b.name))
    .toSorted((a, b) => levelsOrder[b.level] - levelsOrder[a.level]);
});

const { doSubmit: createSecret, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedSecret.value) {
    throw new Error("Unexpected: Can't get secret");
  }

  if (isEditingSecret.value) {
    await apiClient.updateSecret(repo.value.id, selectedSecret.value);
  } else {
    await apiClient.createSecret(repo.value.id, selectedSecret.value);
  }
  notifications.notify({
    title: isEditingSecret.value ? i18n.t('secrets.saved') : i18n.t('secrets.created'),
    type: 'success',
  });
  selectedSecret.value = undefined;
  await resetPage();
});

const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
  await apiClient.deleteSecret(repo.value.id, _secret.name);
  notifications.notify({ title: i18n.t('secrets.deleted'), type: 'success' });
  await resetPage();
});

function editSecret(secret: Secret) {
  selectedSecret.value = cloneDeep(secret);
}

function showAddSecret() {
  selectedSecret.value = cloneDeep(emptySecret);
}
</script>
