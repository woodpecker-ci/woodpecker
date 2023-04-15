<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-color">{{ $t('org.settings.secrets.secrets') }}</h1>
        <p class="text-sm text-color-alt">
          {{ $t('org.settings.secrets.desc') }}
          <DocsLink :topic="$t('org.settings.secrets.secrets')" url="docs/usage/secrets" />
        </p>
      </div>
      <Button
        v-if="selectedSecret"
        class="ml-auto"
        :text="$t('org.settings.secrets.show')"
        start-icon="back"
        @click="selectedSecret = undefined"
      />
      <Button v-else class="ml-auto" :text="$t('org.settings.secrets.add')" start-icon="plus" @click="showAddSecret" />
    </div>

    <SecretList
      v-if="!selectedSecret"
      v-model="secrets"
      i18n-prefix="org.settings.secrets."
      :is-deleting="isDeleting"
      @edit="editSecret"
      @delete="deleteSecret"
    />

    <SecretEdit
      v-else
      v-model="selectedSecret"
      i18n-prefix="org.settings.secrets."
      :is-saving="isSaving"
      @save="createSecret"
      @cancel="selectedSecret = undefined"
    />
  </Panel>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash';
import { computed, defineComponent, inject, Ref, ref } from 'vue';
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
import { Org, Secret, WebhookEvents } from '~/lib/api/types';

const emptySecret = {
  name: '',
  value: '',
  image: [],
  event: [WebhookEvents.Push],
};

export default defineComponent({
  name: 'OrgSecretsTab',

  components: {
    Button,
    Panel,
    DocsLink,
    SecretList,
    SecretEdit,
  },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const i18n = useI18n();

    const org = inject<Ref<Org>>('org');
    const selectedSecret = ref<Partial<Secret>>();
    const isEditingSecret = computed(() => !!selectedSecret.value?.id);

    async function loadSecrets(page: number): Promise<Secret[] | null> {
      if (!org?.value) {
        throw new Error("Unexpected: Can't load org");
      }

      return apiClient.getOrgSecretList(org.value.name, page);
    }

    const { page, data: secrets } = usePagination(loadSecrets, () => !selectedSecret.value);

    const { doSubmit: createSecret, isLoading: isSaving } = useAsyncAction(async () => {
      if (!org?.value) {
        throw new Error("Unexpected: Can't load org");
      }

      if (!selectedSecret.value) {
        throw new Error("Unexpected: Can't get secret");
      }

      if (isEditingSecret.value) {
        await apiClient.updateOrgSecret(org.value.name, selectedSecret.value);
      } else {
        await apiClient.createOrgSecret(org.value.name, selectedSecret.value);
      }
      notifications.notify({
        title: i18n.t(isEditingSecret.value ? 'org.settings.secrets.saved' : 'org.settings.secrets.created'),
        type: 'success',
      });
      selectedSecret.value = undefined;
      secrets.value = [];
      page.value = 1;
    });

    const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
      if (!org?.value) {
        throw new Error("Unexpected: Can't load org");
      }

      await apiClient.deleteOrgSecret(org.value.name, _secret.name);
      notifications.notify({ title: i18n.t('org.settings.secrets.deleted'), type: 'success' });
      secrets.value = [];
      page.value = 1;
    });

    function editSecret(secret: Secret) {
      selectedSecret.value = cloneDeep(secret);
    }

    function showAddSecret() {
      selectedSecret.value = cloneDeep(emptySecret);
    }

    return {
      selectedSecret,
      secrets,
      isDeleting,
      isSaving,
      showAddSecret,
      createSecret,
      editSecret,
      deleteSecret,
    };
  },
});
</script>
