<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-gray-500">{{ $t('repo.settings.secrets.secrets') }}</h1>
        <p class="text-sm text-gray-400 dark:text-gray-600">
          {{ $t('repo.settings.secrets.desc') }}
          <DocsLink url="docs/usage/secrets" />
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

    <div v-if="!selectedSecret" class="space-y-4 text-gray-500">
      <ListItem v-for="secret in secrets" :key="secret.id" class="items-center">
        <span>{{ secret.name }}</span>
        <div class="ml-auto">
          <span
            v-for="event in secret.event"
            :key="event"
            class="bg-gray-400 dark:bg-dark-200 dark:text-gray-500 text-white rounded-md mx-1 py-1 px-2 text-sm"
            >{{ event }}</span
          >
        </div>
        <IconButton icon="edit" class="ml-2 w-8 h-8" @click="selectedSecret = secret" />
        <IconButton
          icon="trash"
          class="ml-2 w-8 h-8 hover:text-red-400"
          :is-loading="isDeleting"
          @click="deleteSecret(secret)"
        />
      </ListItem>

      <div v-if="secrets?.length === 0" class="ml-2">{{ $t('repo.settings.secrets.none') }}</div>
    </div>

    <div v-else class="space-y-4">
      <form @submit.prevent="createSecret">
        <InputField :label="$t('repo.settings.secrets.name')">
          <TextField
            v-model="selectedSecret.name"
            :placeholder="$t('repo.settings.secrets.name')"
            required
            :disabled="isEditingSecret"
          />
        </InputField>

        <InputField :label="$t('repo.settings.secrets.value')">
          <TextField
            v-model="selectedSecret.value"
            :placeholder="$t('repo.settings.secrets.value')"
            :lines="5"
            required
          />
        </InputField>

        <InputField :label="$t('repo.settings.secrets.images.images')">
          <TextField v-model="images" :placeholder="$t('repo.settings.secrets.images.desc')" />
        </InputField>

        <InputField :label="$t('repo.settings.secrets.events.events')">
          <CheckboxesField v-model="selectedSecret.event" :options="secretEventsOptions" />
        </InputField>

        <Button
          :is-loading="isSaving"
          type="submit"
          :text="isEditingSecret ? $t('repo.settings.secrets.save') : $t('repo.settings.secrets.add')"
        />
      </form>
    </div>
  </Panel>
</template>

<script lang="ts">
import { cloneDeep } from 'lodash';
import { computed, defineComponent, inject, onMounted, Ref, ref } from 'vue';

import { useI18n } from 'vue-i18n';
import Button from '~/components/atomic/Button.vue';
import DocsLink from '~/components/atomic/DocsLink.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import { CheckboxOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
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
    ListItem,
    IconButton,
    InputField,
    TextField,
    DocsLink,
    CheckboxesField,
  },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const i18n = useI18n();

    const repo = inject<Ref<Repo>>('repo');
    const secrets = ref<Secret[]>();
    const selectedSecret = ref<Partial<Secret>>();
    const isEditingSecret = computed(() => !!selectedSecret.value?.id);
    const images = computed<string>({
      get() {
        return selectedSecret.value?.image?.join(',') || '';
      },
      set(value) {
        if (selectedSecret.value) {
          selectedSecret.value.image = value
            .split(',')
            .map((s) => s.trim())
            .filter((s) => s !== '');
        }
      },
    });

    async function loadSecrets() {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      secrets.value = await apiClient.getSecretList(repo.value.owner, repo.value.name);
    }

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
      await loadSecrets();
    });

    const { doSubmit: deleteSecret, isLoading: isDeleting } = useAsyncAction(async (_secret: Secret) => {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      await apiClient.deleteSecret(repo.value.owner, repo.value.name, _secret.name);
      notifications.notify({ title: i18n.t('repo.settings.secrets.deleted'), type: 'success' });
      await loadSecrets();
    });

    function showAddSecret() {
      selectedSecret.value = cloneDeep(emptySecret);
    }

    onMounted(async () => {
      await loadSecrets();
    });

    const secretEventsOptions: CheckboxOption[] = [
      { value: WebhookEvents.Push, text: i18n.t('repo.build.event.push') },
      { value: WebhookEvents.Tag, text: i18n.t('repo.build.event.tag') },
      {
        value: WebhookEvents.PullRequest,
        text: i18n.t('repo.build.event.pr'),
        description: i18n.t('repo.settings.secrets.events.pr_warning'),
      },
      { value: WebhookEvents.Deploy, text: i18n.t('repo.build.event.deploy') },
    ];

    return {
      secretEventsOptions,
      selectedSecret,
      secrets,
      images,
      isEditingSecret,
      isSaving,
      isDeleting,
      showAddSecret,
      createSecret,
      deleteSecret,
    };
  },
});
</script>
