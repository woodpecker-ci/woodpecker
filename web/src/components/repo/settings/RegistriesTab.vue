<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-gray-500">Registry credentials</h1>
        <p class="text-sm text-gray-400 dark:text-gray-600">
          Registries credentials can be added to use private images for your pipeline.
          <DocsLink url="docs/usage/registries" />
        </p>
      </div>
      <Button
        v-if="selectedRegistry"
        class="ml-auto"
        start-icon="back"
        text="Show registries"
        @click="selectedRegistry = undefined"
      />
      <Button v-else class="ml-auto" start-icon="plus" text="Add registry" @click="selectedRegistry = {}" />
    </div>

    <div v-if="!selectedRegistry" class="space-y-4 text-gray-500">
      <ListItem v-for="registry in registries" :key="registry.id" class="items-center">
        <span>{{ registry.address }}</span>
        <IconButton icon="edit" class="ml-auto w-8 h-8" @click="selectedRegistry = registry" />
        <IconButton
          icon="trash"
          class="w-8 h-8 hover:text-red-400"
          :is-loading="isDeleting"
          @click="deleteRegistry(registry)"
        />
      </ListItem>

      <div v-if="registries?.length === 0" class="ml-2">There are no registry credentials yet.</div>
    </div>

    <div v-else class="space-y-4">
      <form @submit.prevent="createRegistry">
        <InputField label="Address">
          <!-- TODO: check input field Address is a valid address -->
          <TextField
            v-model="selectedRegistry.address"
            placeholder="Registry Address (e.g. docker.io)"
            required
            :disabled="isEditingRegistry"
          />
        </InputField>

        <InputField label="Username">
          <TextField v-model="selectedRegistry.username" placeholder="Username" required />
        </InputField>

        <InputField label="Password">
          <TextField v-model="selectedRegistry.password" placeholder="Password" required />
        </InputField>

        <Button type="submit" :is-loading="isSaving" :text="isEditingRegistry ? 'Save registy' : 'Add registry'" />
      </form>
    </div>
  </Panel>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, Ref, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import DocsLink from '~/components/atomic/DocsLink.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { Repo } from '~/lib/api/types';
import { Registry } from '~/lib/api/types/registry';

export default defineComponent({
  name: 'RegistriesTab',

  components: {
    Button,
    Panel,
    ListItem,
    IconButton,
    InputField,
    TextField,
    DocsLink,
  },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');
    const registries = ref<Registry[]>();
    const selectedRegistry = ref<Partial<Registry>>();
    const isEditingRegistry = computed(() => !!selectedRegistry.value?.id);

    async function loadRegistries() {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      registries.value = await apiClient.getRegistryList(repo.value.owner, repo.value.name);
    }

    const { doSubmit: createRegistry, isLoading: isSaving } = useAsyncAction(async () => {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      if (!selectedRegistry.value) {
        throw new Error("Unexpected: Can't get registry");
      }

      if (isEditingRegistry.value) {
        await apiClient.updateRegistry(repo.value.owner, repo.value.name, selectedRegistry.value);
      } else {
        await apiClient.createRegistry(repo.value.owner, repo.value.name, selectedRegistry.value);
      }
      notifications.notify({ title: 'Registry credentials created', type: 'success' });
      selectedRegistry.value = undefined;
      await loadRegistries();
    });

    const { doSubmit: deleteRegistry, isLoading: isDeleting } = useAsyncAction(async (_registry: Registry) => {
      if (!repo?.value) {
        throw new Error("Unexpected: Can't load repo");
      }

      const registryAddress = encodeURIComponent(_registry.address);
      await apiClient.deleteRegistry(repo.value.owner, repo.value.name, registryAddress);
      notifications.notify({ title: 'Registry credentials deleted', type: 'success' });
      await loadRegistries();
    });

    onMounted(async () => {
      await loadRegistries();
    });

    return { selectedRegistry, registries, isEditingRegistry, isSaving, isDeleting, createRegistry, deleteRegistry };
  },
});
</script>
