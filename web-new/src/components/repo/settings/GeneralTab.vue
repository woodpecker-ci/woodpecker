<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center">
      <h1 class="text-xl ml-2">General</h1>
    </div>

    <div class="flex flex-col">
      <InputField label="Pipeline path">
        <TextField v-model="settings.config_file" />
      </InputField>

      <InputField label="Repository hooks">
        <CheckboxesField v-model="settings.repository_hooks" :options="repositoryHooksOptions" />
      </InputField>

      <InputField label="Project settings">
        <CheckboxesField v-model="settings.repository_hooks" :options="projectSettingsOptions" />
      </InputField>

      <InputField label="Project visibility">
        <RadioField v-model="settings.visibility" :options="projectVisibilityOptions" />
      </InputField>

      <InputField label="Timeout">
        <TextField v-model="settings.timeout" />
        <span>minutes</span>
      </InputField>

      <Button class="mx-auto bg-green hover:bg-lime-600 text-white" text="Save settings" @click="deleteRepo" />
    </div>

    <div class="flex mt-8 pt-4 border-t-1">
      <Button class="mx-auto bg-red-500 hover:bg-red-400 text-white" text="Delete repository" @click="deleteRepo" />
    </div>
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, ref, Ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo, SecretEvents } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';
import { useRouter } from 'vue-router';
import useNotifications from '~/compositions/useNotifications';
import Panel from '~/components/layout/Panel.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import RadioField from '~/components/form/RadioField.vue';
import { CheckboxOption, RadioOption } from '~/components/form/form.types';
import CheckboxesField from '~/components/form/CheckboxesField.vue';

const repositoryHooksOptions: CheckboxOption[] = [
  { value: SecretEvents.Push, text: 'Push' },
  { value: SecretEvents.Tag, text: 'Tag' },
  { value: SecretEvents.PullRequest, text: 'Pull Request' },
  { value: SecretEvents.Deploy, text: 'Deploy' },
];

const projectSettingsOptions: CheckboxOption[] = [
  { value: 'protected', text: 'Protected' },
  { value: 'trusted', text: 'Trusted' },
];

const projectVisibilityOptions: RadioOption[] = [
  { value: 'public', text: 'Public' },
  { value: 'private', text: 'Private' },
  { value: 'internal', text: 'Internal' },
];

export default defineComponent({
  name: 'GeneralTab',

  components: { Button, Panel, InputField, TextField, RadioField, CheckboxesField },

  setup() {
    const apiClient = useApiClient();
    const router = useRouter();
    const notifications = useNotifications();

    const repo = inject<Ref<Repo>>('repo');
    const settings = ref({
      config_file: repo?.value.config_file,
      timeout: repo?.value.timeout,
      visibility: repo?.value.visibility,
      repository_hooks: [],
    });

    async function deleteRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      if (!confirm('All data will be lost after this action!!!\n\nDo you really want to procceed?')) {
        return;
      }

      await apiClient.deleteRepo(repo.value.owner, repo.value.name);
      notifications.notify({ title: 'Repository deleted', type: 'success' });
      await router.replace({ name: 'repos' });
    }

    return { deleteRepo, repo, settings, projectVisibilityOptions, projectSettingsOptions, repositoryHooksOptions };
  },
});
</script>
