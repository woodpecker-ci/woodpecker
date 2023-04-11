<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-color">{{ $t('admin.settings.secrets.secrets') }}</h1>
        <p class="text-sm text-color-alt">
          {{ $t('admin.settings.secrets.desc') }}
          <DocsLink :topic="$t('admin.settings.secrets.secrets')" url="docs/usage/secrets" />
        </p>
        <Warning :text="$t('admin.settings.secrets.warning')" />
      </div>
      <Button
        v-if="selectedSecret"
        class="ml-auto"
        :text="$t('admin.settings.secrets.show')"
        start-icon="back"
        @click="selectedSecret = undefined"
      />
      <Button
        v-else
        class="ml-auto"
        :text="$t('admin.settings.secrets.add')"
        start-icon="plus"
        @click="showAddSecret"
      />
    </div>

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
  </Panel>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash';
import { computed, defineComponent, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import DocsLink from '~/components/atomic/DocsLink.vue';
import Warning from '~/components/atomic/Warning.vue';
import Panel from '~/components/layout/Panel.vue';
import SecretEdit from '~/components/secrets/SecretEdit.vue';
import SecretList from '~/components/secrets/SecretList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { PaginatedList } from '~/compositions/usePaginate';
import { Secret, WebhookEvents } from '~/lib/api/types';

const emptySecret = {
  name: '',
  value: '',
  image: [],
  event: [WebhookEvents.Push],
};

export default defineComponent({
  name: 'AdminSecretsTab',

  components: {
    Button,
    Panel,
    DocsLink,
    SecretList,
    SecretEdit,
    Warning,
  },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const i18n = useI18n();

    const secrets = ref<Secret[]>([]);
    const selectedSecret = ref<Partial<Secret>>();
    const isEditingSecret = computed(() => !!selectedSecret.value?.id);

    async function loadSecrets(page: number): Promise<boolean> {
      const sec = await apiClient.getGlobalSecretList(page);
      if (page === 1 && sec !== null) {
        secrets.value = sec;
      } else if (sec !== null) {
        secrets.value?.push(...sec);
      }
      return sec !== null && sec.length !== 0;
    }

    const list = new PaginatedList(loadSecrets, () => !selectedSecret.value);

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
      list.reset(true);
    });

    const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
      await apiClient.deleteGlobalSecret(_secret.name);
      notifications.notify({ title: i18n.t('admin.settings.secrets.deleted'), type: 'success' });
      list.reset(true);
    });

    function editSecret(secret: Secret) {
      selectedSecret.value = cloneDeep(secret);
    }

    function showAddSecret() {
      selectedSecret.value = cloneDeep(emptySecret);
    }

    onMounted(() => {
      list.init();
    });

    onUnmounted(() => {
      list.clear();
    });

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
