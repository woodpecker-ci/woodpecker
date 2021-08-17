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
          <span v-for="event in secret.event" :key="event" class="bg-gray-400 text-white rounded-md mx-1 px-1 py-0.5">{{
            event
          }}</span>
        </div>
        <IconButton icon="trash" @click="deleteSecret(secret)" />
      </ListItem>
    </div>

    <div v-else class="space-y-4">
      <form @submit.prevent="createSecret">
        <input v-model="secret.name" type="text" placeholder="Name" required />
        <input v-model="secret.value" type="text" placeholder="Value" required />

        <div v-for="secretEvent in SecretEvents" :key="secretEvent">
          <input
            type="checkbox"
            @click="clickSecretEvent(secretEvent)"
            :id="`event-${secretEvent}`"
            :value="secretEvent"
            :checked="secret.event?.includes(secretEvent)"
          />
          <label :for="`event-${secretEvent}`">{{ secretEvent }}</label>
        </div>

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

const emptySecret = {
  name: '',
  value: '',
  image: [],
  event: [SecretEvents.Push],
};

export default defineComponent({
  name: 'SecretsTab',

  components: {
    Button,
    Panel,
    ListItem,
    IconButton,
    Icon,
  },

  setup() {
    const apiClient = useApiClient();
    const { notify } = useNotifications();

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
      notify({ title: 'Secret created', type: 'success' });
      showAddSecret.value = false;
      secret.value = { ...emptySecret };
      await loadSecrets();
    }

    function clickSecretEvent(secretEvent: SecretEvents) {
      let events = secret.value.event || [];

      if (events.includes(secretEvent)) {
        events = events.filter((s) => s !== secretEvent);
      } else {
        events.push(secretEvent);
      }

      secret.value = { ...secret.value, event: events };
    }

    async function deleteSecret(secret: Secret) {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      await apiClient.deleteSecret(repo.value.owner, repo.value.name, secret.name);
      notify({ title: 'Secret deleted', type: 'success' });
      await loadSecrets();
    }

    onMounted(async () => {
      await loadSecrets();
    });

    return { SecretEvents, secret, secrets, showAddSecret, createSecret, clickSecretEvent, deleteSecret };
  },
});
</script>
