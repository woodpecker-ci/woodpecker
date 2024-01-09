<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-wp-background-100">
      <div class="ml-2">
        <h1 class="text-xl text-wp-text-100">{{ $t('user.settings.secrets.secrets') }}</h1>
        <p class="text-sm text-wp-text-alt-100">
          {{ $t('user.settings.secrets.desc') }}
          <DocsLink :topic="$t('user.settings.secrets.secrets')" url="docs/usage/secrets" />
        </p>
      </div>
      <Button
        v-if="selectedSecret"
        class="ml-auto"
        :text="$t('user.settings.secrets.show')"
        start-icon="back"
        @click="selectedSecret = undefined"
      />
      <Button v-else class="ml-auto" :text="$t('user.settings.secrets.add')" start-icon="plus" @click="showAddSecret" />
    </div>

    <SecretList
      v-if="!selectedSecret"
      v-model="secrets"
      i18n-prefix="user.settings.secrets."
      :is-deleting="isDeleting"
      @edit="editSecret"
      @delete="deleteSecret"
    />

    <SecretEdit
      v-else
      v-model="selectedSecret"
      i18n-prefix="user.settings.secrets."
      :is-saving="isSaving"
      @save="createSecret"
      @cancel="selectedSecret = undefined"
    />
  </Panel>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import DocsLink from '~/components/atomic/DocsLink.vue';
import Panel from '~/components/layout/Panel.vue';
import SecretEdit from '~/components/secrets/SecretEdit.vue';
import SecretList from '~/components/secrets/SecretList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { Secret, WebhookEvents } from '~/lib/api/types';

const emptySecret: Partial<Secret> = {
  name: '',
  value: '',
  images: [],
  events: [WebhookEvents.Push],
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const { user } = useAuthentication();
if (!user) {
  throw new Error('Unexpected: Unauthenticated');
}
const selectedSecret = ref<Partial<Secret>>();
const isEditingSecret = computed(() => !!selectedSecret.value?.id);

async function loadSecrets(page: number): Promise<Secret[] | null> {
  if (!user) {
    throw new Error('Unexpected: Unauthenticated');
  }

  return apiClient.getOrgSecretList(user.org_id, page);
}

const { resetPage, data: secrets } = usePagination(loadSecrets, () => !selectedSecret.value);

const { doSubmit: createSecret, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedSecret.value) {
    throw new Error("Unexpected: Can't get secret");
  }

  if (isEditingSecret.value) {
    await apiClient.updateOrgSecret(user.org_id, selectedSecret.value);
  } else {
    await apiClient.createOrgSecret(user.org_id, selectedSecret.value);
  }
  notifications.notify({
    title: i18n.t(isEditingSecret.value ? 'user.settings.secrets.saved' : 'user.settings.secrets.created'),
    type: 'success',
  });
  selectedSecret.value = undefined;
  resetPage();
});

const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
  await apiClient.deleteOrgSecret(user.org_id, _secret.name);
  notifications.notify({ title: i18n.t('user.settings.secrets.deleted'), type: 'success' });
  resetPage();
});

function editSecret(secret: Secret) {
  selectedSecret.value = cloneDeep(secret);
}

function showAddSecret() {
  selectedSecret.value = cloneDeep(emptySecret);
}
</script>
