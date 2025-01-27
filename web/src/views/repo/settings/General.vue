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
          <div class="flex flex-col gap-2">
            <div v-for="image in repoSettings.netrc_trusted" :key="image" class="flex gap-2">
              <TextField :id="id" :model-value="image" disabled />
              <Button type="button" color="gray" start-icon="trash" @click="removeImage(image)" />
            </div>
            <div class="flex gap-2">
              <TextField :id="id" v-model="newImage" @keydown.enter.prevent="addNewImage" />
              <Button type="button" color="gray" start-icon="plus" @click="addNewImage" />
            </div>
          </div>
        </template>
        <template #description>
          <p class="text-sm">
            {{ $t('repo.settings.general.netrc_only_trusted.desc') }}
          </p>
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
          <p class="text-sm">
            {{ $t('require_approval.desc') }}
          </p>
        </template>
      </InputField>

      <InputField
        v-if="repoSettings.require_approval !== RepoRequireApproval.None"
        :label="$t('require_approval.allowed_users.allowed_users')"
      >
        <template #default="{ id }">
          <div class="flex flex-col gap-2">
            <div v-for="user in repoSettings.approval_allowed_users" :key="user" class="flex gap-2">
              <TextField :id="id" :model-value="user" disabled />
              <Button type="button" color="gray" start-icon="trash" @click="removeUser(user)" />
            </div>
            <div class="flex gap-2">
              <TextField :id="id" v-model="newUser" @keydown.enter.prevent="addNewUser" />
              <Button type="button" color="gray" start-icon="plus" @click="addNewUser" />
            </div>
          </div>
        </template>
        <template #description>
          <p class="text-sm">
            {{ $t('require_approval.allowed_users.desc') }}
          </p>
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
          <NumberField :id="id" v-model="repoSettings.timeout" class="w-24" />
          <span class="ml-4 text-wp-text-alt-100">{{ $t('repo.settings.general.timeout.minutes') }}</span>
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
            :placeholder="$t('repo.settings.general.pipeline_path.default')"
          />
        </template>
        <template #description>
          <i18n-t keypath="repo.settings.general.pipeline_path.desc" tag="p" class="text-sm text-wp-text-alt-100">
            <span class="code-box-inline">{{ $t('repo.settings.general.pipeline_path.desc_path_example') }}</span>
            <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
            <span class="code-box-inline">/</span>
          </i18n-t>
        </template>
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
  </Settings>
</template>

<script lang="ts" setup>
import { inject, onMounted, ref } from 'vue';
import type { Ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import CheckboxesField from '~/components/form/CheckboxesField.vue';
import type { CheckboxOption, RadioOption } from '~/components/form/form.types';
import InputField from '~/components/form/InputField.vue';
import NumberField from '~/components/form/NumberField.vue';
import RadioField from '~/components/form/RadioField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useAuthentication from '~/compositions/useAuthentication';
import useNotifications from '~/compositions/useNotifications';
import { RepoRequireApproval, RepoVisibility, WebhookEvents } from '~/lib/api/types';
import type { Repo, RepoSettings } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';

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

const newImage = ref('');
function addNewImage() {
  if (!newImage.value) {
    return;
  }
  repoSettings.value?.netrc_trusted.push(newImage.value);
  newImage.value = '';
}
function removeImage(image: string) {
  if (!repoSettings.value) {
    throw new Error('Unexpected: repoSettings should be set');
  }

  repoSettings.value.netrc_trusted = repoSettings.value.netrc_trusted.filter((i) => i !== image);
}

const newUser = ref('');
function addNewUser() {
  if (!newUser.value) {
    return;
  }
  repoSettings.value?.approval_allowed_users.push(newUser.value);
  newUser.value = '';
}
function removeUser(user: string) {
  if (!repoSettings.value) {
    throw new Error('Unexpected: repoSettings should be set');
  }

  repoSettings.value.approval_allowed_users = repoSettings.value.approval_allowed_users.filter((i) => i !== user);
}
</script>
