<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-color">{{ $t('extensions') }}</h1>
    </div>

    <div class="flex flex-col">
      <span class="text-color font-bold">{{ $t('extensions_signatures_public_key') }}</span>
      <span class="text-color">{{ $t('extensions_signatures_public_key_description') }}</span>
      <CodeBox>{{ signaturePublicKey }}</CodeBox>
    </div>

    <div class="flex flex-col mt-4 border-t-1 dark:border-gray-600">
      <form @submit.prevent="saveExtensions">
        <InputField :label="$t('secrets_extension_endpoint')" docs-url="docs/usage/extensions/secrets-extension">
          <TextField
            v-model="extensions.secret_extension_endpoint"
            :placeholder="$t('secrets_extension_endpoint_placeholder')"
          />
          <template #description>
            <p class="text-sm">{{ $t('secrets_extension_alpha_state') }}</p>
          </template>
        </InputField>

        <InputField :label="$t('config_extension_endpoint')" docs-url="docs/usage/extensions/configuration-extension">
          <TextField
            v-model="extensions.config_extension_endpoint"
            :placeholder="$t('extension_endpoint_placeholder')"
          />
        </InputField>

        <Button :is-loading="isSaving" color="green" type="submit" :text="$t('save')" />
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
  secret_extension_endpoint: repo.value.secret_extension_endpoint,
  config_extension_endpoint: repo.value.config_extension_endpoint,
});

const { doSubmit: saveExtensions, isLoading: isSaving } = useAsyncAction(async () => {
  await apiClient.updateRepo(repo.value.id, extensions.value);

  // await loadRepo();
  notifications.notify({ title: i18n.t('extensions_configuration_saved'), type: 'success' });
});
</script>
