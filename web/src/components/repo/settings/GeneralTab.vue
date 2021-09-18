<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">General</h1>
    </div>

    <div v-if="repoSettings" class="flex flex-col">
      <InputField label="Pipeline path">
        <TextField
          v-model="repoSettings.config_file"
          class="max-w-124"
          placeholder="By default: .woodpecker/*.yml -> .woodpecker.yml -> .drone.yml"
        />
        <template #description>
          <p class="text-sm text-gray-600">
            Path to your pipeline config (for example <span class="bg-gray-300 rounded-md px-1">my/path/</span>).
            Folders should end with a <span class="bg-gray-300 rounded-md px-1">/</span>.
            <a :href="`${docsUrl}docs/usage/project-settings#pipeline-path`" target="_blank" class="text-blue-500"
              >(i)</a
            >
          </p>
        </template>
      </InputField>

      <InputField label="Project settings">
        <Checkbox v-model="repoSettings.allow_pr" label="Allow Pull Request" />
        <Checkbox v-model="repoSettings.gated" label="Protected" />
        <Checkbox v-model="repoSettings.trusted" label="Trusted" />
      </InputField>

      <InputField label="Project visibility">
        <RadioField v-model="repoSettings.visibility" :options="projectVisibilityOptions" />
      </InputField>

      <InputField label="Timeout">
        <div class="flex items-center">
          <NumberField v-model="repoSettings.timeout" class="w-24" />
          <span class="ml-4 text-gray-600">minutes</span>
        </div>
      </InputField>

      <Button class="mr-auto" color="green" text="Save settings" @click="saveRepoSettings" />
    </div>
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import { RadioOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import NumberField from '~/components/form/NumberField.vue';
import RadioField from '~/components/form/RadioField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { Repo, RepoSettings, RepoVisibility } from '~/lib/api/types';
import RepoStore from '~/store/repos';

const projectVisibilityOptions: RadioOption[] = [
  { value: RepoVisibility.Public, text: 'Public' },
  { value: RepoVisibility.Private, text: 'Private' },
  { value: RepoVisibility.Internal, text: 'Internal' },
];

export default defineComponent({
  name: 'GeneralTab',

  components: { Button, Panel, InputField, TextField, RadioField, NumberField, Checkbox },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const repoStore = RepoStore();

    const repo = inject<Ref<Repo>>('repo');
    const repoSettings = ref<RepoSettings>();
    const docsUrl = window.WOODPECKER_DOCS;

    function loadRepoSettings() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      repoSettings.value = {
        config_file: repo.value.config_file,
        timeout: repo.value.timeout,
        visibility: repo.value.visibility,
        gated: repo.value.gated,
        trusted: repo.value.trusted,
        allow_pr: repo.value.allow_pr,
      };
    }

    async function loadRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await repoStore.loadRepo(repo.value.owner, repo.value.name);
      loadRepoSettings();
    }

    async function saveRepoSettings() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      if (!repoSettings.value) {
        throw new Error('Unexpected: Repo-Settings should be set');
      }

      await apiClient.updateRepo(repo.value.owner, repo.value.name, repoSettings.value);
      await loadRepo();
      notifications.notify({ title: 'Repository settings updated', type: 'success' });
    }

    onMounted(() => {
      loadRepoSettings();
    });

    return {
      docsUrl,
      repoSettings,
      saveRepoSettings,
      projectVisibilityOptions,
    };
  },
});
</script>
