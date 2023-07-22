<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-wp-background-100">
      <h1 class="text-xl ml-2 text-wp-text-100">{{ $t('repo.settings.general.general') }}</h1>
    </div>

    <form v-if="repoSettings" class="flex flex-col" @submit.prevent="saveRepoSettings">
      <InputField
        docs-url="docs/usage/project-settings#pipeline-path"
        :label="$t('repo.settings.general.pipeline_path.path')"
      >
        <TextField
          v-model="repoSettings.config_file"
          :placeholder="$t('repo.settings.general.pipeline_path.default')"
        />
        <template #description>
          <i18n-t keypath="repo.settings.general.pipeline_path.desc" tag="p" class="text-sm text-wp-text-alt-100">
            <span class="bg-wp-code-200 text-wp-code-text-100 rounded-md px-1">{{
              $t('repo.settings.general.pipeline_path.desc_path_example')
            }}</span>
            <span class="bg-wp-code-200 text-wp-code-text-100 rounded-md px-1">/</span>
          </i18n-t>
        </template>
      </InputField>

      <InputField
        docs-url="docs/usage/project-settings#project-settings-1"
        :label="$t('repo.settings.general.project')"
      >
        <Checkbox
          v-model="repoSettings.allow_pr"
          :label="$t('repo.settings.general.allow_pr.allow')"
          :description="$t('repo.settings.general.allow_pr.desc')"
        />
        <Checkbox
          v-model="repoSettings.gated"
          :label="$t('repo.settings.general.protected.protected')"
          :description="$t('repo.settings.general.protected.desc')"
        />
        <Checkbox
          v-model="repoSettings.netrc_only_trusted"
          :label="$t('repo.settings.general.netrc_only_trusted.netrc_only_trusted')"
          :description="$t('repo.settings.general.netrc_only_trusted.desc')"
        />
        <Checkbox
          v-if="user?.admin"
          v-model="repoSettings.trusted"
          :label="$t('repo.settings.general.trusted.trusted')"
          :description="$t('repo.settings.general.trusted.desc')"
        />
      </InputField>

      <InputField
        docs-url="docs/usage/project-settings#project-visibility"
        :label="$t('repo.settings.general.visibility.visibility')"
      >
        <RadioField v-model="repoSettings.visibility" :options="projectVisibilityOptions" />
      </InputField>

      <InputField docs-url="docs/usage/project-settings#timeout" :label="$t('repo.settings.general.timeout.timeout')">
        <div class="flex items-center">
          <NumberField v-model="repoSettings.timeout" class="w-24" />
          <span class="ml-4 text-wp-gray-600">{{ $t('repo.settings.general.timeout.minutes') }}</span>
        </div>
      </InputField>

      <InputField
        docs-url="docs/usage/project-settings#cancel-previous-pipelines"
        :label="$t('repo.settings.general.cancel_prev.cancel')"
      >
        <CheckboxesField
          v-model="repoSettings.cancel_previous_pipeline_events"
          :options="cancelPreviousPipelineEventsOptions"
        />
        <template #description>
          <p class="text-sm">
            {{ $t('repo.settings.general.cancel_prev.desc') }}
          </p>
        </template>
      </InputField>

      <Button
        type="submit"
        class="mr-auto"
        color="green"
        :is-loading="isSaving"
        :text="$t('repo.settings.general.save')"
      />
    </form>
  </Panel>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref } from 'vue';
import { useI18n } from 'vue-i18n';

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
import { useRepoStore } from '~/store/repos';

export default defineComponent({
  name: 'GeneralTab',

  components: { Button, Panel, InputField, TextField, RadioField, NumberField, Checkbox, CheckboxesField },

  setup() {
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const { user } = useAuthentication();
    const repoStore = useRepoStore();
    const i18n = useI18n();

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
        cancel_previous_pipeline_events: repo.value.cancel_previous_pipeline_events || [],
        netrc_only_trusted: repo.value.netrc_only_trusted,
      };
    }

    async function loadRepo() {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      await repoStore.loadRepo(repo.value.id);
      loadRepoSettings();
    }

    const { doSubmit: saveRepoSettings, isLoading: isSaving } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo should be set');
      }

      if (!repoSettings.value) {
        throw new Error('Unexpected: Repo-Settings should be set');
      }

      await apiClient.updateRepo(repo.value.id, repoSettings.value);
      await loadRepo();
      notifications.notify({ title: i18n.t('repo.settings.general.success'), type: 'success' });
    });

    onMounted(() => {
      loadRepoSettings();
    });

    const projectVisibilityOptions: RadioOption[] = [
      {
        value: RepoVisibility.Public,
        text: i18n.t('repo.settings.general.visibility.public.public'),
        description: i18n.t('repo.settings.general.visibility.public.desc'),
      },
      {
        value: RepoVisibility.Internal,
        text: i18n.t('repo.settings.general.visibility.internal.internal'),
        description: i18n.t('repo.settings.general.visibility.internal.desc'),
      },
      {
        value: RepoVisibility.Private,
        text: i18n.t('repo.settings.general.visibility.private.private'),
        description: i18n.t('repo.settings.general.visibility.private.desc'),
      },
    ];

    const cancelPreviousPipelineEventsOptions: CheckboxOption[] = [
      { value: WebhookEvents.Push, text: i18n.t('repo.pipeline.event.push') },
      { value: WebhookEvents.Tag, text: i18n.t('repo.pipeline.event.tag') },
      {
        value: WebhookEvents.PullRequest,
        text: i18n.t('repo.pipeline.event.pr'),
      },
      { value: WebhookEvents.Deploy, text: i18n.t('repo.pipeline.event.deploy') },
    ];

    return {
      user,
      repoSettings,
      isSaving,
      saveRepoSettings,
      projectVisibilityOptions,
      cancelPreviousPipelineEventsOptions,
    };
  },
});
</script>
