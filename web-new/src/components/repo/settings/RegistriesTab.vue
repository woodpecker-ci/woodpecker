<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">Registries</h1>
      <Button v-if="showAddRegistry" class="ml-auto" @click="showAddRegistry = false" text="Show registries" />
      <Button v-else class="ml-auto" @click="showAddRegistry = true" text="Add registry" />
    </div>

    <div v-if="!showAddRegistry" class="space-y-4">
      <ListItem v-for="registry in registries" :key="registry.id" class="items-center">
        <span>{{ registry.address }}</span>
        <IconButton class="ml-auto" icon="trash" @click="deleteRegistry(registry)" />
      </ListItem>
    </div>

    <div v-else class="space-y-4">
      <form @submit.prevent="createRegistry">
        <input v-model="registry.address" type="text" placeholder="Address" required />
        <input v-model="registry.username" type="text" placeholder="Username" required />
        <input v-model="registry.password" type="text" placeholder="Password" required />

        <Button type="submit" text="Add registry" />
      </form>
    </div>
  </Panel>
</template>

<script lang="ts">
import { ref, defineComponent, inject, onMounted, Ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import useNotifications from '~/compositions/useNotifications';
import Panel from '~/components/layout/Panel.vue';
import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import Icon from '~/components/atomic/Icon.vue';
import { Registry } from '~/lib/api/types/registry';

export default defineComponent({
  name: 'RegistriesTab',

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
    const registries = ref<Registry[]>();
    const showAddRegistry = ref(false);
    const registry = ref<Partial<Registry>>({});

    async function loadRegistries() {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      registries.value = await apiClient.getRegistryList(repo.value.owner, repo.value.name);
    }

    async function createRegistry() {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      await apiClient.createRegistry(repo.value.owner, repo.value.name, registry.value);
      notify({ title: 'Registry created', type: 'success' });
      showAddRegistry.value = false;
      registry.value = {};
      await loadRegistries();
    }

    async function deleteRegistry(registry: Registry) {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      await apiClient.deleteRegistry(repo.value.owner, repo.value.name, registry.address);
      notify({ title: 'Registry deleted', type: 'success' });
      await loadRegistries();
    }

    onMounted(async () => {
      await loadRegistries();
    });

    return { registry, registries, showAddRegistry, createRegistry, deleteRegistry };
  },
});
</script>
