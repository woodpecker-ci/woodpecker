<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">Secrets</h1>
      <Button v-if="showAddSecret" class="ml-auto" @click="showAddSecret = false" text="Show secrets" />
      <Button v-else class="ml-auto" @click="showAddSecret = true" text="Add secret" />
    </div>

    <div v-if="!showAddSecret" class="space-y-4">
      <ListItem v-for="secret in secrets" :key="secret.id" class="items-center">
        <span>{{ secret.name }}</span>
        <div class="ml-auto">
          <span
            v-for="event in secret.event"
            :key="event"
            class="bg-gray-400 text-white rounded-md mx-1 py-1 px-2 text-sm"
            >{{ event }}</span
          >
        </div>
        <IconButton icon="trash" class="ml-2 w-6 h-6 hover:text-red-400" @click="deleteSecret(secret)" />
      </ListItem>

      <div v-if="secrets?.length === 0">There are no secrets yet.</div>
    </div>

    <div v-else class="space-y-4">
      <form @submit.prevent="createSecret">
        <InputField label="Name">
          <TextField v-model="secret.name" placeholder="Name" required />
        </InputField>

        <InputField label="Value">
          <TextField v-model="secret.value" placeholder="Value" required />
        </InputField>

        <InputField label="Available at following events">
          <CheckboxesField :options="secretEventsOptions" v-model="secret.event" />
        </InputField>

        <Button type="submit" text="Add secret" />
      </form>
    </div>
  </Panel>
</template>

<script lang="ts">
import { ref, defineComponent, inject, onMounted, Ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo, Secret, SecretEvents } from '~/lib/api/types';
import useNotifications from '~/compositions/useNotifications';
import Panel from '~/components/layout/Panel.vue';
import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import Icon from '~/components/atomic/Icon.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import { CheckboxOption } from '~/components/form/form.types';

const emptySecret = {
  name: '',
  value: '',
  image: [],
  event: [SecretEvents.Push],
};

const secretEventsOptions: CheckboxOption[] = [
  { value: SecretEvents.Push, text: 'Push' },
  { value: SecretEvents.Tag, text: 'Tag' },
  { value: SecretEvents.PullRequest, text: 'Pull Request' },
  { value: SecretEvents.Deploy, text: 'Deploy' },
];

export default defineComponent({
  name: 'SecretsTab',

  components: {
    Button,
    Panel,
    ListItem,
    IconButton,
    Icon,
    InputField,
    TextField,
    CheckboxesField,
  },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');
    const secrets = ref<Secret[]>();
    const showAddSecret = ref(false);
    const secret = ref<Partial<Secret>>({ ...emptySecret });

    async function loadSecrets() {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      secrets.value = await apiClient.getSecretList(repo.value.owner, repo.value.name);
    }

    async function createSecret() {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      await apiClient.createSecret(repo.value.owner, repo.value.name, secret.value);
      notifications.notify({ title: 'Secret created', type: 'success' });
      showAddSecret.value = false;
      secret.value = { ...emptySecret };
      await loadSecrets();
    }

    async function deleteSecret(secret: Secret) {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      await apiClient.deleteSecret(repo.value.owner, repo.value.name, secret.name);
      notifications.notify({ title: 'Secret deleted', type: 'success' });
      await loadSecrets();
    }

    onMounted(async () => {
      await loadSecrets();
    });

    return { secretEventsOptions, secret, secrets, showAddSecret, createSecret, deleteSecret };
  },
});
</script>
