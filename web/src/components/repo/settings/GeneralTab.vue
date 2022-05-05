<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <h1 class="text-xl ml-2 text-gray-500">General</h1>
    </div>

    <div v-if="repoSettings" class="flex flex-col">
      <InputField label="Pipeline path" docs-url="docs/usage/project-settings#pipeline-path">
        <TextField
          v-model="repoSettings.config_file"
          class="max-w-124"
          placeholder="By default: .woodpecker/*.yml -> .woodpecker.yml -> .drone.yml"
        />
        <template #description>
          <p class="text-sm text-gray-400 dark:text-gray-600">
            Path to your pipeline config (for example
            <span class="bg-gray-300 dark:bg-dark-100 rounded-md px-1">my/path/</span>). Folders should end with a
            <span class="bg-gray-300 dark:bg-dark-100 rounded-md px-1">/</span>.
          </p>
        </template>
      </InputField>

      <InputField label="Project settings" docs-url="docs/usage/project-settings#project-settings-1">
        <Checkbox
          v-model="repoSettings.allow_pr"
          label="Allow Pull Request"
          description="Pipelines can run on pull requests."
        />
        <Checkbox
          v-model="repoSettings.gated"
          label="Protected"
          description="Every pipeline needs to be approved before being executed."
        />
        <Checkbox
          v-if="user?.admin"
          v-model="repoSettings.trusted"
          label="Trusted"
          description="Underlying pipeline containers get access to escalated capabilities like mounting volumes."
        />
      </InputField>

      <InputField label="Project visibility" docs-url="docs/usage/project-settings#project-visibility">
        <RadioField v-model="repoSettings.visibility" :options="projectVisibilityOptions" />
      </InputField>

      <InputField label="Timeout" docs-url="docs/usage/project-settings#timeout">
        <div class="flex items-center">
          <NumberField v-model="repoSettings.timeout" class="w-24" />
          <span class="ml-4 text-gray-600">minutes</span>
        </div>
      </InputField>

      <InputField label="Cancel previous pipelines" docs-url="docs/usage/project-settings#project-settings-1">
        <CheckboxesField
          v-model="repoSettings.cancel_previous_build_events"
          :options="cancelPreviousBuildEventsOptions"
        />
        <template #description>
          <p class="text-sm text-gray-400 dark:text-gray-600">
            Enable to cancel running pipelines of the same event and context before starting the newly triggered one.
          </p>
        </template>
      </InputField>

      <Button class="mr-auto" color="green" text="Save settings" :is-loading="isSaving" @click="saveRepoSettings" />
    </div>
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import { CheckboxOption, RadioOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import NumberField from '~/components/form/NumberField.vue';
import RadioField from '~/components/form/RadioField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';
import { Repo, RepoSettings, RepoVisibility, WebhookEvents } from '~/lib/api/types';
import RepoStore from '~/store/repos';

const projectVisibilityOptions: RadioOption[] = [
  {
    value: RepoVisibility.Public,
    text: 'Public',
    description: 'Every user can see your project without being logged in.',
  },
  {
    value: RepoVisibility.Private,
    text: 'Private',
    description: 'Only authenticated users of the Woodpecker instance can see this project.',
  },
  {
    value: RepoVisibility.Internal,
    text: 'Internal',
    description: 'Only you and other owners of the repository can see this project.',
  },
];

const cancelPreviousBuildEventsOptions: CheckboxOption[] = [
  { value: WebhookEvents.Push, text: 'Push' },
  { value: WebhookEvents.Tag, text: 'Tag' },
  {
    value: WebhookEvents.PullRequest,
    text: 'Pull Request',
  },
  { value: WebhookEvents.Deploy, text: 'Deploy' },
];

export default defineComponent({
  name: 'GeneralTab',

  components: { Button, Panel, InputField, TextField, RadioField, NumberField, Checkbox, CheckboxesField },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const { user } = useAuthentication();
    const repoStore = RepoStore();

    const repo = inject<Ref<Repo>>('repo');
    const repoSettings = ref<RepoSettings>();

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
        cancel_previous_build_events: repo.value.cancel_previous_build_events || [],
      };
    }

    async function loadRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await repoStore.loadRepo(repo.value.owner, repo.value.name);
      loadRepoSettings();
    }

    const { doSubmit: saveRepoSettings, isLoading: isSaving } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      if (!repoSettings.value) {
        throw new Error('Unexpected: Repo-Settings should be set');
      }

      await apiClient.updateRepo(repo.value.owner, repo.value.name, repoSettings.value);
      await loadRepo();
      notifications.notify({ title: 'Repository settings updated', type: 'success' });
    });

    onMounted(() => {
      loadRepoSettings();
    });

    return {
      user,
      repoSettings,
      isSaving,
      saveRepoSettings,
      projectVisibilityOptions,
      cancelPreviousBuildEventsOptions,
    };
  },
});
</script>
