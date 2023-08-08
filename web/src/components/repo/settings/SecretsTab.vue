<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-wp-background-100">
      <div class="ml-2">
        <h1 class="text-xl text-wp-text-100">{{ $t('repo.settings.secrets.secrets') }}</h1>
        <p class="text-sm text-wp-text-alt-100">
          {{ $t('repo.settings.secrets.desc') }}
          <DocsLink :topic="$t('repo.settings.secrets.secrets')" url="docs/usage/secrets" />
        </p>
      </div>
      <Button
        v-if="selectedSecret"
        class="ml-auto"
        :text="$t('repo.settings.secrets.show')"
        start-icon="back"
        @click="selectedSecret = undefined"
      />
      <Button v-else class="ml-auto" :text="$t('repo.settings.secrets.add')" start-icon="plus" @click="showAddSecret" />
    </div>

    <SecretList
      v-if="!selectedSecret"
      v-model="secrets"
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
  </Panel>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, inject, Ref, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import DocsLink from '~/components/atomic/DocsLink.vue';
import Panel from '~/components/layout/Panel.vue';
import SecretEdit from '~/components/secrets/SecretEdit.vue';
import SecretList from '~/components/secrets/SecretList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { Repo, Secret, WebhookEvents } from '~/lib/api/types';

const emptySecret = {
  name: '',
  value: '',
  image: [],
  event: [WebhookEvents.Push],
};

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

const repo = inject<Ref<Repo>>('repo');
const selectedSecret = ref<Partial<Secret>>();
const isEditingSecret = computed(() => !!selectedSecret.value?.id);

async function loadSecrets(page: number): Promise<Secret[] | null> {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  return apiClient.getSecretList(repo.value.id, page);
}

const { resetPage, data: secrets } = usePagination(loadSecrets, () => !selectedSecret.value);

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
  resetPage();
});

const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
  if (!repo?.value) {
    throw new Error("Unexpected: Can't load repo");
  }

  await apiClient.deleteSecret(repo.value.id, _secret.name);
  notifications.notify({ title: i18n.t('repo.settings.secrets.deleted'), type: 'success' });
  resetPage();
});

function editSecret(secret: Secret) {
  selectedSecret.value = cloneDeep(secret);
}

function showAddSecret() {
  selectedSecret.value = cloneDeep(emptySecret);
}
</script>
