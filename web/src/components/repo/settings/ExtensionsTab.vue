<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-color">{{ $t('repo.settings.extensions.extensions') }}</h1>
    </div>

    <div class="flex flex-col">
      <span class="text-color font-bold">{{ $t('repo.settings.extensions.signatures_public_key') }}</span>
      <span class="text-color">{{ $t('repo.settings.extensions.signatures_public_key_desc') }}</span>
      <CodeBox>{{ signaturePublicKey }}</CodeBox>
    </div>

    <div class="flex flex-col mt-4 border-t-1 dark:border-gray-600">
      <form @submit.prevent="saveExtensions">
        <InputField
          :label="$t('repo.settings.extensions.secrets_endpoint')"
          docs-url="docs/usage/extensions/secret-extension"
        >
          <TextField
            v-model="extensions.secret_endpoint"
            :placeholder="$t('repo.settings.extensions.secrets_endpoint_placeholder')"
          />
        </InputField>

        <InputField
          :label="$t('repo.settings.extensions.registries_endpoint')"
          docs-url="docs/usage/extensions/registry-extension"
        >
          <TextField
            v-model="extensions.registry_endpoint"
            :placeholder="$t('repo.settings.extensions.registries_endpoint_placeholder')"
          />
        </InputField>

        <InputField
          :label="$t('repo.settings.extensions.config_endpoint')"
          docs-url="docs/usage/extensions/configuration-extension"
        >
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
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { ExtensionSettings, Repo } from '~/lib/api/types';

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

const extensions = ref<ExtensionSettings>({
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
