<template>
  <Settings
    :title="$t('secrets.secrets')"
    :description="$t('admin.settings.secrets.desc')"
    docs-url="docs/usage/secrets"
  >
    <template #headerActions>
      <Button v-if="selectedSecret" :text="$t('secrets.show')" start-icon="back" @click="selectedSecret = undefined" />
      <Button v-else :text="$t('secrets.add')" start-icon="plus" @click="showAddSecret" />
    </template>

    <template #headerEnd>
      <Warning class="mt-4 text-sm" :text="$t('admin.settings.secrets.warning')" />
    </template>

    <SecretList
      v-if="!selectedSecret"
      v-model="secrets"
      :is-deleting="isDeleting"
      :loading="loading"
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
import Warning from '~/components/atomic/Warning.vue';
import Settings from '~/components/layout/Settings.vue';
import SecretEdit from '~/components/secrets/SecretEdit.vue';
import SecretList from '~/components/secrets/SecretList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Secret } from '~/lib/api/types';
import { WebhookEvents } from '~/lib/api/types';

const emptySecret: Partial<Secret> = {
  name: '',
  value: '',
  images: [],
  events: [WebhookEvents.Push],
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const selectedSecret = ref<Partial<Secret>>();
const isEditingSecret = computed(() => !!selectedSecret.value?.id);

async function loadSecrets(page: number): Promise<Secret[] | null> {
  return apiClient.getGlobalSecretList({ page });
}

const { resetPage, data: secrets, loading } = usePagination(loadSecrets, () => !selectedSecret.value);

const { doSubmit: createSecret, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedSecret.value) {
    throw new Error("Unexpected: Can't get secret");
  }

  if (isEditingSecret.value) {
    await apiClient.updateGlobalSecret(selectedSecret.value);
  } else {
    await apiClient.createGlobalSecret(selectedSecret.value);
  }
  notifications.notify({
    title: isEditingSecret.value ? i18n.t('secrets.saved') : i18n.t('secrets.created'),
    type: 'success',
  });
  selectedSecret.value = undefined;
  await resetPage();
});

const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
  await apiClient.deleteGlobalSecret(_secret.name);
  notifications.notify({ title: i18n.t('secrets.deleted'), type: 'success' });
  await resetPage();
});

function editSecret(secret: Secret) {
  selectedSecret.value = cloneDeep(secret);
}

function showAddSecret() {
  selectedSecret.value = cloneDeep(emptySecret);
}

useWPTitle(computed(() => [i18n.t('secrets.secrets'), i18n.t('admin.settings.settings')]));
</script>
