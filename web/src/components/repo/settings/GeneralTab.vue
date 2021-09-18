<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">General</h1>
    </div>

    <div v-if="repoSettings" class="flex flex-col">
      <InputField label="Pipeline path">
        <TextField v-model="repoSettings.config_file" class="max-w-124" />
      </InputField>

      <InputField label="Repository hooks">
        <Checkbox v-model="repoSettings.allow_push" label="Push" />
        <Checkbox v-model="repoSettings.allow_pr" label="Pull Request" />
        <Checkbox v-model="repoSettings.allow_tags" label="Tag" />
        <Checkbox v-model="repoSettings.allow_deploys" label="Deploy" />
      </InputField>

      <InputField label="Project settings">
        <Checkbox v-model="repoSettings.gated" label="Protected" />
        <Checkbox v-model="repoSettings.trusted" label="Trusted" />
      </InputField>

      <InputField label="Project visibility">
        <RadioField v-model="repoSettings.visibility" :options="projectVisibilityOptions" />
      </InputField>

      <InputField label="Timeout">
        <div class="flex items-center">
          <NumberField v-model="repoSettings.timeout" class="w-24" />
          <span class="ml-4">minutes</span>
        </div>
      </InputField>

      <Button class="mr-auto bg-green hover:bg-lime-600 text-white" text="Save settings" @click="saveRepoSettings" />
    </div>

    <div class="flex flex-col mt-8 pt-4 border-t-1">
      <span class="text-xl">Actions</span>
      <Button
        class="mr-auto mt-4 bg-red-500 hover:bg-red-400 text-white"
        text="Delete repository"
        @click="deleteRepo"
      />
    </div>
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref } from 'vue';
import { useRouter } from 'vue-router';

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
    const router = useRouter();
    const notifications = useNotifications();
    const repoStore = RepoStore();

    const repo = inject<Ref<Repo>>('repo');
    const repoSettings = ref<RepoSettings>();

    function loadRepoSettings() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      repoSettings.value = {
        config_file: repo.value.config_file,
        fallback: repo.value.fallback,
        timeout: repo.value.timeout,
        visibility: repo.value.visibility,
        gated: repo.value.gated,
        trusted: repo.value.trusted,
        allow_push: repo.value.allow_push,
        allow_pr: repo.value.allow_pr,
        allow_tags: repo.value.allow_tags,
        allow_deploys: repo.value.allow_deploys,
      };
    }

    async function loadRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await repoStore.loadRepo(repo.value.owner, repo.value.name);
      loadRepoSettings();
    }

    async function deleteRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      // TODO use proper dialog
      // eslint-disable-next-line no-alert, no-restricted-globals
      if (!confirm('All data will be lost after this action!!!\n\nDo you really want to procceed?')) {
        return;
      }

      await apiClient.deleteRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository deleted', type: 'success' });
      await router.replace({ name: 'repos' });
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
      deleteRepo,
      repoSettings,
      saveRepoSettings,
      projectVisibilityOptions,
    };
  },
});
</script>
