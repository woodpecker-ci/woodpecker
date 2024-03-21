<template>
  <Settings
    :title="$t('repo.settings.secrets.secrets')"
    :desc="$t('repo.settings.secrets.desc')"
    docs-url="docs/usage/secrets"
  >
    <template #titleActions>
      <Button
        v-if="selectedSecret"
        :text="$t('repo.settings.secrets.show')"
        start-icon="back"
        @click="selectedSecret = undefined"
      />
      <Button v-else :text="$t('repo.settings.secrets.add')" start-icon="plus" @click="showAddSecret" />
    </template>

    <SecretList
      v-if="!selectedSecret"
      :model-value="secrets"
      i18n-prefix="repo.settings.secrets."
      :is-deleting="isDeleting"
      @edit="editSecret"
      @delete="deleteSecret"
    />

    <SecretEdit
      v-else
      v-model="selectedSecret"
      i18n-prefix="repo.settings.secrets."
      :is-saving="isSaving"
      @save="createSecret"
      @cancel="selectedSecret = undefined"
    />
  </Settings>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, inject, Ref, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Settings from '~/components/layout/Settings.vue';
import SecretEdit from '~/components/secrets/SecretEdit.vue';
import SecretList from '~/components/secrets/SecretList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { Repo, Secret, WebhookEvents } from '~/lib/api/types';

const emptySecret: Partial<Secret> = {
  name: '',
  value: '',
  images: [],
  events: [WebhookEvents.Push],
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const repo = inject<Ref<Repo>>('repo');
const selectedSecret = ref<Partial<Secret>>();
const isEditingSecret = computed(() => !!selectedSecret.value?.id);

async function loadSecrets(page: number, level: 'repo' | 'org' | 'global'): Promise<Secret[] | null> {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  switch (level) {
    case 'repo':
      return apiClient.getSecretList(repo.value.id, page);
    case 'org':
      return apiClient.getOrgSecretList(repo.value.org_id, page);
    case 'global':
      return apiClient.getGlobalSecretList(page);
    default:
      throw new Error(`Unexpected level: ${level}`);
  }
}

const { resetPage, data: _secrets } = usePagination(loadSecrets, () => !selectedSecret.value, {
  each: ['repo', 'org', 'global'],
});
const secrets = computed(() => {
  const secretsList: Record<string, Secret & { edit?: boolean; level: 'repo' | 'org' | 'global' }> = {};

  // eslint-disable-next-line no-restricted-syntax
  for (const level of ['repo', 'org', 'global']) {
    // eslint-disable-next-line no-restricted-syntax
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
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  if (!selectedSecret.value) {
    throw new Error("Unexpected: Can't get secret");
  }

  if (isEditingSecret.value) {
    await apiClient.updateSecret(repo.value.id, selectedSecret.value);
  } else {
    await apiClient.createSecret(repo.value.id, selectedSecret.value);
  }
  notifications.notify({
    title: i18n.t(isEditingSecret.value ? 'repo.settings.secrets.saved' : 'repo.settings.secrets.created'),
    type: 'success',
  });
  selectedSecret.value = undefined;
  await resetPage();
});

const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  await apiClient.deleteSecret(repo.value.id, _secret.name);
  notifications.notify({ title: i18n.t('repo.settings.secrets.deleted'), type: 'success' });
  await resetPage();
});

function editSecret(secret: Secret) {
  selectedSecret.value = cloneDeep(secret);
}

function showAddSecret() {
  selectedSecret.value = cloneDeep(emptySecret);
}
</script>
