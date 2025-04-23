<template>
  <Settings :title="$t('secrets.secrets')" :description="$t('org.settings.secrets.desc')" docs-url="docs/usage/secrets">
    <template #headerActions>
      <Button v-if="selectedSecret" :text="$t('secrets.show')" start-icon="back" @click="selectedSecret = undefined" />
      <Button v-else :text="$t('secrets.add')" start-icon="plus" @click="showAddSecret" />
    </template>

    <SecretList
      v-if="!selectedSecret"
      v-model="secrets"
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

const org = requiredInject('org');
const selectedSecret = ref<Partial<Secret>>();
const isEditingSecret = computed(() => !!selectedSecret.value?.id);

async function loadSecrets(page: number): Promise<Secret[] | null> {
  return apiClient.getOrgSecretList(org.value.id, { page });
}

const { resetPage, data: secrets } = usePagination(loadSecrets, () => !selectedSecret.value);

const { doSubmit: createSecret, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedSecret.value) {
    throw new Error("Unexpected: Can't get secret");
  }

  if (isEditingSecret.value) {
    await apiClient.updateOrgSecret(org.value.id, selectedSecret.value);
  } else {
    await apiClient.createOrgSecret(org.value.id, selectedSecret.value);
  }
  notifications.notify({
    title: isEditingSecret.value ? i18n.t('secrets.saved') : i18n.t('secrets.created'),
    type: 'success',
  });
  selectedSecret.value = undefined;
  resetPage();
});

const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
  await apiClient.deleteOrgSecret(org.value.id, _secret.name);
  notifications.notify({ title: i18n.t('secrets.deleted'), type: 'success' });
  resetPage();
});

function editSecret(secret: Secret) {
  selectedSecret.value = cloneDeep(secret);
}

function showAddSecret() {
  selectedSecret.value = cloneDeep(emptySecret);
}
</script>
