<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-color">{{ $t('repo.settings.secrets.secrets') }}</h1>
        <p class="text-sm text-color-alt">
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

<script lang="ts">
import { cloneDeep } from 'lodash';
import { computed, defineComponent, inject, onMounted, onUnmounted, Ref, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import DocsLink from '~/components/atomic/DocsLink.vue';
import Panel from '~/components/layout/Panel.vue';
import SecretEdit from '~/components/secrets/SecretEdit.vue';
import SecretList from '~/components/secrets/SecretList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { PaginatedList } from '~/compositions/usePaginate';
import { Repo, Secret, WebhookEvents } from '~/lib/api/types';

const emptySecret = {
  name: '',
  value: '',
  image: [],
  event: [WebhookEvents.Push],
};

export default defineComponent({
  name: 'SecretsTab',

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

    const repo = inject<Ref<Repo>>('repo');
    const secrets = ref<Secret[]>([]);
    const selectedSecret = ref<Partial<Secret>>();
    const isEditingSecret = computed(() => !!selectedSecret.value?.id);

    async function loadSecrets(page: number): Promise<boolean> {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      const sec = await apiClient.getSecretList(repo.value.owner, repo.value.name, page);
      if (page === 1 && sec !== null) {
        secrets.value = sec;
      } else if (sec !== null) {
        secrets.value?.push(...sec);
      }
      return sec !== null && sec.length !== 0;
    }

    const list = new PaginatedList(loadSecrets, () => !selectedSecret.value);

    const { doSubmit: createSecret, isLoading: isSaving } = useAsyncAction(async () => {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      if (!selectedSecret.value) {
        throw new Error("Unexpected: Can't get secret");
      }

      if (isEditingSecret.value) {
        await apiClient.updateSecret(repo.value.owner, repo.value.name, selectedSecret.value);
      } else {
        await apiClient.createSecret(repo.value.owner, repo.value.name, selectedSecret.value);
      }
      notifications.notify({
        title: i18n.t(isEditingSecret.value ? 'repo.settings.secrets.saved' : 'repo.settings.secrets.created'),
        type: 'success',
      });
      selectedSecret.value = undefined;
      list.reset(true);
    });

    const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      await apiClient.deleteSecret(repo.value.owner, repo.value.name, _secret.name);
      notifications.notify({ title: i18n.t('repo.settings.secrets.deleted'), type: 'success' });
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
