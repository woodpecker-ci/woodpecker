<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-color">{{ $t('repo.settings.extensions.extensions') }}</h1>
    </div>

    <div>
      <span class="text-color font-bold">{{ $t('repo.settings.extensions.signatures_public_key') }}</span>
      <div class="token-box mt-2">{{ signaturePublicKey }}</div>
    </div>

    <div class="flex flex-col mt-4 border-t-1 dark:border-gray-600">
      <form @submit.prevent="saveExtensions">
        <InputField :label="$t('repo.settings.extensions.secrets_endpoint')">
          <TextField
            v-model="extensions.secret_endpoint"
            :placeholder="$t('repo.settings.extensions.secrets_endpoint_placeholder')"
          />
        </InputField>

        <InputField :label="$t('repo.settings.extensions.registries_endpoint')">
          <TextField
            v-model="extensions.registry_endpoint"
            :placeholder="$t('repo.settings.extensions.registries_endpoint_placeholder')"
          />
        </InputField>

        <InputField :label="$t('repo.settings.extensions.config_endpoint')">
          <TextField
            v-model="extensions.config_endpoint"
            :placeholder="$t('repo.settings.extensions.config_endpoint_placeholder')"
          />
        </InputField>

        <Button :is-loading="isSaving" type="submit" :text="$t('repo.settings.extensions.save_extensions')" />
      </form>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { inject, onMounted, Ref, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { Repo } from '~/lib/api/types';

const i18n = useI18n();

const apiClient = useApiClient();
const notifications = useNotifications();

const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Missing repo');
}

const signaturePublicKey = ref<string>();

onMounted(async () => {
  signaturePublicKey.value = await apiClient.getSignaturePublicKey();
});

const extensions = ref<Pick<Repo, 'config_endpoint' | 'registry_endpoint' | 'secret_endpoint'>>({
  secret_endpoint: repo.value.secret_endpoint,
  registry_endpoint: repo.value.registry_endpoint,
  config_endpoint: repo.value.config_endpoint,
});

const { doSubmit: saveExtensions, isLoading: isSaving } = useAsyncAction(async () => {
  await apiClient.updateRepo(repo.value.owner, repo.value.name, extensions.value);

  // await loadRepo();
  notifications.notify({ title: i18n.t('repo.settings.extensions.success'), type: 'success' });
});
</script>

<style scoped>
.token-box {
  @apply bg-gray-500 p-2 rounded-md text-white break-words dark:bg-dark-400 dark:text-gray-400;
  white-space: pre-wrap;
}
</style>
