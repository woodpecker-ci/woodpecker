<template>
  <Settings :title="$t('repo.settings.general.project')">
    <form v-if="repoSettings" class="flex flex-col" @submit.prevent="saveRepoSettings">
      <InputField
        docs-url="docs/usage/project-settings#project-settings-1"
        :label="$t('repo.settings.general.general')"
      >
        <Checkbox
          v-model="repoSettings.allow_pr"
          :label="$t('repo.settings.general.allow_pr.allow')"
          :description="$t('repo.settings.general.allow_pr.desc')"
        />
        <Checkbox
          v-model="repoSettings.allow_deploy"
          :label="$t('repo.settings.general.allow_deploy.allow')"
          :description="$t('repo.settings.general.allow_deploy.desc')"
        />
      </InputField>

      <InputField
        :label="$t('repo.settings.general.netrc_only_trusted.netrc_only_trusted')"
        docs-url="docs/usage/project-settings#custom-trusted-clone-plugins"
      >
        <template #default="{ id }">
          <ListEditor
            :id="id"
            ref="netrcTrustedEditor"
            v-model="repoSettings.netrc_trusted"
            :placeholder="$t('repo.settings.general.netrc_only_trusted.placeholder')"
          />
        </template>
        <template #description>
          {{ $t('repo.settings.general.netrc_only_trusted.desc') }}
        </template>
      </InputField>

      <InputField
        v-if="user?.admin"
        docs-url="docs/usage/project-settings#project-settings-1"
        :label="$t('repo.settings.general.trusted.trusted')"
      >
        <Checkbox
          v-model="repoSettings.trusted.network"
          :label="$t('repo.settings.general.trusted.network.network')"
          :description="$t('repo.settings.general.trusted.network.desc')"
        />
        <Checkbox
          v-model="repoSettings.trusted.volumes"
          :label="$t('repo.settings.general.trusted.volumes.volumes')"
          :description="$t('repo.settings.general.trusted.volumes.desc')"
        />
        <Checkbox
          v-model="repoSettings.trusted.security"
          :label="$t('repo.settings.general.trusted.security.security')"
          :description="$t('repo.settings.general.trusted.security.desc')"
        />
      </InputField>

      <InputField :label="$t('require_approval.require_approval_for')">
        <RadioField
          v-model="repoSettings.require_approval"
          :options="[
            {
              value: RepoRequireApproval.None,
              text: $t('require_approval.none'),
              description: $t('require_approval.none_desc'),
            },
            {
              value: RepoRequireApproval.Forks,
              text: $t('require_approval.forks'),
            },
            {
              value: RepoRequireApproval.PullRequests,
              text: $t('require_approval.pull_requests'),
            },
            {
              value: RepoRequireApproval.AllEvents,
              text: $t('require_approval.all_events'),
            },
          ]"
        />
        <template #description>
          {{ $t('require_approval.desc') }}
        </template>
      </InputField>

      <InputField
        v-if="repoSettings.require_approval !== RepoRequireApproval.None"
        :label="$t('require_approval.allowed_users.allowed_users')"
      >
        <template #default="{ id }">
          <ListEditor
            :id="id"
            ref="approvalAllowedUsersEditor"
            v-model="repoSettings.approval_allowed_users"
            :placeholder="$t('username')"
          />
        </template>
        <template #description>
          {{ $t('require_approval.allowed_users.desc') }}
        </template>
      </InputField>

      <InputField docs-url="docs/usage/project-settings#project-visibility" :label="$t('repo.visibility.visibility')">
        <RadioField v-model="repoSettings.visibility" :options="projectVisibilityOptions" />
      </InputField>

      <InputField
        v-slot="{ id }"
        docs-url="docs/usage/project-settings#timeout"
        :label="$t('repo.settings.general.timeout.timeout')"
      >
        <div class="flex items-center">
          <NumberField
            :id="id"
            v-model="repoSettings.timeout"
            :placeholder="$t('repo.settings.general.timeout.timeout')"
            class="w-24"
          />
          <span class="text-wp-text-alt-100 ml-4">{{ $t('repo.settings.general.timeout.minutes') }}</span>
        </div>
      </InputField>

      <InputField
        docs-url="docs/usage/project-settings#pipeline-path"
        :label="$t('repo.settings.general.pipeline_path.path')"
      >
        <template #default="{ id }">
          <TextField
            :id="id"
            v-model="repoSettings.config_file"
            :placeholder="
              $t('repo.settings.general.pipeline_path.by_default', { paths: defaultConfigPaths.join(' ➔ ') })
            "
          />
        </template>

        <!-- eslint-disable @intlify/vue-i18n/no-raw-text -->
        <template #description>
          <i18n-t keypath="repo.settings.general.pipeline_path.desc">
            <span class="code-box-inline">{{ $t('repo.settings.general.pipeline_path.desc_path_example') }}</span>
            <span class="code-box-inline">/</span>
          </i18n-t>
        </template>
        <!-- eslint-enable @intlify/vue-i18n/no-raw-text -->
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
          {{ $t('repo.settings.general.cancel_prev.desc') }}
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
  </Settings>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref, useTemplateRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import type { CheckboxOption, RadioOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import ListEditor from '~/components/form/ListEditor.vue';
import NumberField from '~/components/form/NumberField.vue';
import RadioField from '~/components/form/RadioField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import { requiredInject } from '~/compositions/useInjectProvide';
import useNotifications from '~/compositions/useNotifications';
import { useWPTitle } from '~/compositions/useWPTitle';
import { RepoRequireApproval, RepoVisibility, WebhookEvents } from '~/lib/api/types';
import type { RepoSettings } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';

const apiClient = useApiClient();
const notifications = useNotifications();
const { user } = useAuthentication();
const repoStore = useRepoStore();
const i18n = useI18n();
const { defaultConfigPaths } = useConfig();

const repo = requiredInject('repo');
const repoSettings = ref<RepoSettings>();

const netrcTrustedEditor = useTemplateRef<InstanceType<typeof ListEditor>>('netrcTrustedEditor');
const approvalAllowedUsersEditor = useTemplateRef<InstanceType<typeof ListEditor>>('approvalAllowedUsersEditor');

function loadRepoSettings() {
  repoSettings.value = {
    config_file: repo.value.config_file,
    timeout: repo.value.timeout,
    visibility: repo.value.visibility,
    require_approval: repo.value.require_approval,
    trusted: repo.value.trusted,
    approval_allowed_users: repo.value.approval_allowed_users || [],
    allow_pr: repo.value.allow_pr,
    allow_deploy: repo.value.allow_deploy,
    cancel_previous_pipeline_events: repo.value.cancel_previous_pipeline_events || [],
    netrc_trusted: repo.value.netrc_trusted || [],
  };
}

async function loadRepo() {
  await repoStore.loadRepo(repo.value.id);
  loadRepoSettings();
}

const { doSubmit: saveRepoSettings, isLoading: isSaving } = useAsyncAction(async () => {
  if (!repoSettings.value) {
    throw new Error('Unexpected: Repo-Settings should be set');
  }

  // an entry the user typed without confirming it should still be saved
  netrcTrustedEditor.value?.commitPendingItem();
  approvalAllowedUsersEditor.value?.commitPendingItem();

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
    text: i18n.t('repo.visibility.public.public'),
    description: i18n.t('repo.visibility.public.desc'),
  },
  {
    value: RepoVisibility.Internal,
    text: i18n.t('repo.visibility.internal.internal'),
    description: i18n.t('repo.visibility.internal.desc'),
  },
  {
    value: RepoVisibility.Private,
    text: i18n.t('repo.visibility.private.private'),
    description: i18n.t('repo.visibility.private.desc'),
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

useWPTitle(computed(() => [i18n.t('repo.settings.general.project'), repo.value.full_name]));
</script>
