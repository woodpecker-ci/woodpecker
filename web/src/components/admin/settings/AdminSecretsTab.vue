<template>
  <Settings
    :title="$t('admin.settings.secrets.secrets')"
    :desc="$t('admin.settings.secrets.desc')"
    docs-url="docs/usage/secrets"
    :warning="$t('admin.settings.secrets.warning')"
  >
    <template #titleActions>
      <Button
        v-if="selectedSecret"
        :text="$t('admin.settings.secrets.show')"
        start-icon="back"
        @click="selectedSecret = undefined"
      />
      <Button v-else :text="$t('admin.settings.secrets.add')" start-icon="plus" @click="showAddSecret" />
    </template>

    <SecretList
      v-if="!selectedSecret"
      v-model="secrets"
      i18n-prefix="admin.settings.secrets."
      :is-deleting="isDeleting"
      @edit="editSecret"
      @delete="deleteSecret"
    />

    <SecretEdit
      v-else
      v-model="selectedSecret"
      i18n-prefix="admin.settings.secrets."
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
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { Secret, WebhookEvents } from '~/lib/api/types';

const emptySecret = {
  name: '',
  value: '',
  image: [],
  event: [WebhookEvents.Push],
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const selectedSecret = ref<Partial<Secret>>();
const isEditingSecret = computed(() => !!selectedSecret.value?.id);

async function loadSecrets(page: number): Promise<Secret[] | null> {
  return apiClient.getGlobalSecretList(page);
}

const { resetPage, data: secrets } = usePagination(loadSecrets, () => !selectedSecret.value);

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
    title: i18n.t(isEditingSecret.value ? 'admin.settings.secrets.saved' : 'admin.settings.secrets.created'),
    type: 'success',
  });
  selectedSecret.value = undefined;
  resetPage();
});

const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
  await apiClient.deleteGlobalSecret(_secret.name);
  notifications.notify({ title: i18n.t('admin.settings.secrets.deleted'), type: 'success' });
  resetPage();
});

function editSecret(secret: Secret) {
  selectedSecret.value = cloneDeep(secret);
}

function showAddSecret() {
  selectedSecret.value = cloneDeep(emptySecret);
}
</script>
