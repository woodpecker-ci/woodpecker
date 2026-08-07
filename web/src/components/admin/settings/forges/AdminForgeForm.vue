<template>
  <form @submit.prevent="submit">
    <Warning v-if="!isNew && forge.id === 1" :text="$t('forge_managed_by_env')" />

    <InputField v-slot="{ id }" :label="$t('forge_type')">
      <SelectField :id="id" v-model="forgeType" :options="forgeTypeOptions" required />
    </InputField>

    <InputField v-if="forge.type !== 'bitbucket'" v-slot="{ id }" :label="$t('url')">
      <TextField :id="id" v-model="forge.url" :placeholder="$t('url')" required />
    </InputField>

    <InputField :label="$t('oauth_redirect_url')">
      <template #default="{ id }">
        <TextField :id="id" class="mt-2" :model-value="redirectUri" :label="$t('oauth_redirect_url')" disabled />
      </template>
      <template #description>
        {{ $t('use_this_redirect_url_to_create') }}
        <i18n-t v-if="forge.type !== 'addon'" keypath="developer_settings_to_create" tag="span">
          <a rel="noopener noreferrer" :href="oauthAppForgeUrl" target="_blank" class="underline">{{
            $t('developer_settings')
          }}</a>
        </i18n-t>
      </template>
    </InputField>

    <template v-if="forge.type !== 'addon'">
      <InputField v-slot="{ id }" :label="$t('oauth_client_id')">
        <TextField :id="id" v-model="forge.client" required :placeholder="$t('oauth_client_id')" />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('oauth_client_secret')">
        <TextField
          :id="id"
          v-model="forge.oauth_client_secret"
          :placeholder="isNew ? $t('oauth_client_secret') : $t('leave_empty_to_keep_current_value')"
          :required="isNew"
        />
      </InputField>
    </template>

    <template v-else>
      <InputField v-slot="{ id }" :label="$t('executable')">
        <p>{{ $t('executable_desc') }}</p>
        <TextField
          :id="id"
          :placeholder="$t('executable')"
          :model-value="getAdditionalOptions('addon', 'executable')"
          @update:model-value="setAdditionalOptions('addon', 'executable', $event)"
        />
      </InputField>
    </template>

    <Panel
      v-if="forge.type !== 'bitbucket'"
      collapsable
      collapsed-by-default
      :title="$t('advanced_options')"
      class="mb-4"
    >
      <InputField v-slot="{ id }" :label="$t('oauth_host')">
        <TextField :id="id" v-model="forge.oauth_host" :placeholder="$t('public_url_for_oauth_if', [forge.url])" />
      </InputField>

      <template v-if="forge.type === 'github'">
        <InputField :label="$t('merge_ref')">
          <Checkbox
            :label="$t('merge_ref_desc')"
            :model-value="getAdditionalOptions('github', 'merge-ref') ?? false"
            @update:model-value="setAdditionalOptions('github', 'merge-ref', $event)"
          />
        </InputField>

        <InputField :label="$t('public_only')">
          <Checkbox
            :label="$t('public_only_desc')"
            :model-value="getAdditionalOptions('github', 'public-only') ?? false"
            @update:model-value="setAdditionalOptions('github', 'public-only', $event)"
          />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('github_app_id')">
          <p>
            {{ $t('github_app_id_desc') }}
            <template v-if="githubAppSettingsUrl">
              <a :href="githubAppSettingsUrl" target="_blank" rel="noopener noreferrer" class="underline">{{
                $t('github_app_create_new')
              }}</a>
            </template>
          </p>
          <TextField
            :id="id"
            :placeholder="$t('github_app_id')"
            :model-value="getAdditionalOptions('github', 'app-id')"
            :required="!!getAdditionalOptions('github', 'app-private-key')"
            @update:model-value="setAdditionalOptions('github', 'app-id', $event)"
          />
        </InputField>
        <InputField v-slot="{ id }" :label="$t('github_app_private_key')">
          <p>{{ $t('github_app_private_key_desc') }}</p>
          <TextField
            :id="id"
            :lines="5"
            :placeholder="hasStoredAppKey ? $t('leave_empty_to_keep_current_value') : $t('github_app_private_key')"
            :model-value="getAdditionalOptions('github', 'app-private-key')"
            :required="!hasStoredAppKey && !!getAdditionalOptions('github', 'app-id')"
            @update:model-value="setAdditionalOptions('github', 'app-private-key', $event)"
          />
        </InputField>
        <InputField v-if="getAdditionalOptions('github', 'app-id')" :label="$t('github_app_clone_token_scope')">
          <Checkbox
            :label="$t('github_app_clone_token_scope_desc')"
            :model-value="getAdditionalOptions('github', 'app-clone-token-scope') === 'installation'"
            @update:model-value="
              setAdditionalOptions('github', 'app-clone-token-scope', $event ? 'installation' : 'repo')
            "
          />
        </InputField>
        <InputField
          v-if="!isNew && forge.id && getAdditionalOptions('github', 'app-id')"
          :label="$t('github_app_check')"
        >
          <p>{{ $t('github_app_check_desc') }}</p>
          <div class="flex items-center gap-2">
            <Button type="button" :text="$t('github_app_check')" :is-loading="isCheckingApp" @click="checkGithubApp" />
            <span v-if="appHealth" :class="appHealth.healthy ? 'text-wp-state-ok-100' : 'text-wp-state-error-100'">
              {{
                appHealth.healthy
                  ? $t('github_app_check_healthy', {
                      app: appHealth.app_name,
                      installations: appHealth.installations,
                    })
                  : $t('github_app_check_unhealthy', { error: appHealth.error })
              }}
            </span>
          </div>
        </InputField>
      </template>
      <template v-if="forge.type === 'bitbucket-dc'">
        <InputField v-slot="{ id }" :label="$t('git_username')">
          <p>{{ $t('git_username_desc') }}</p>
          <TextField
            :id="id"
            :placeholder="$t('git_username')"
            :model-value="getAdditionalOptions('bitbucket-dc', 'git-username')"
            @update:model-value="setAdditionalOptions('bitbucket-dc', 'git-username', $event)"
          />
        </InputField>
        <InputField v-slot="{ id }" :label="$t('git_password')">
          <p>{{ $t('git_password_desc') }}</p>
          <TextField
            :id="id"
            :placeholder="$t('git_password')"
            :model-value="getAdditionalOptions('bitbucket-dc', 'git-password')"
            @update:model-value="setAdditionalOptions('bitbucket-dc', 'git-password', $event)"
          />
        </InputField>
      </template>

      <InputField :label="$t('skip_verify')">
        <Checkbox
          :label="$t('skip_verify_desc')"
          :model-value="forge.skip_verify || false"
          @update:model-value="forge!.skip_verify = $event"
        />
      </InputField>
    </Panel>

    <div class="flex gap-2">
      <Button :text="$t('cancel')" :to="{ name: 'admin-settings-forges' }" />

      <Button :is-loading="isSaving" type="submit" color="green" :text="isNew ? $t('add') : $t('save')" />
    </div>
  </form>
</template>

<script lang="ts" setup>
import { computed, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Warning from '~/components/atomic/Warning.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useApiClient from '~/compositions/useApiClient';
import useConfig from '~/compositions/useConfig';
import type { Forge, ForgeAppHealth, ForgeType } from '~/lib/api/types';

defineProps<{
  isNew?: boolean;
  isSaving?: boolean;
}>();

const emit = defineEmits<{
  (e: 'submit'): void;
}>();

const { t } = useI18n();

const config = useConfig();

const forge = defineModel<Partial<Forge>>('forge', {
  required: true,
});

// Define forge type options
const forgeTypeOptions = [
  { value: 'github', text: t('github') },
  { value: 'gitlab', text: t('gitlab') },
  { value: 'gitea', text: t('gitea') },
  { value: 'forgejo', text: t('forgejo') },
  { value: 'bitbucket', text: t('bitbucket') },
  { value: 'bitbucket-dc', text: t('bitbucket_dc') },
  { value: 'addon', text: t('addon') },
];

// Function to get default URL for a forge type
function getDefaultUrl(forgeType: ForgeType): string {
  switch (forgeType) {
    case 'github':
      return 'github.com';
    case 'gitlab':
      return 'gitlab.com';
    case 'bitbucket':
      return 'bitbucket.org';
    default:
      return '';
  }
}

// Initialize forge type to have a default value (first option)
if (!forge.value.type) {
  const defaultType = forgeTypeOptions[0].value as ForgeType;
  forge.value.type = defaultType;
  forge.value.url = forge.value.url || getDefaultUrl(defaultType);
}

// Initialize forge type to have a default value
if (!forge.value.type) {
  forge.value.type = 'github';
}

interface GitHubAdditionOptions {
  'merge-ref'?: boolean;
  'public-only'?: boolean;
  'app-id'?: string;
  'app-private-key'?: string;
  'app-clone-token-scope'?: string;
}

interface BitbucketAdditionOptions {
  'git-username'?: string;
  'git-password'?: string;
}

interface AddonAdditionOptions {
  executable?: string;
}

function getAdditionalOptions<T extends keyof GitHubAdditionOptions>(
  forgeType: 'github',
  key: T,
): GitHubAdditionOptions[T];
// eslint-disable-next-line no-redeclare
function getAdditionalOptions<T extends keyof BitbucketAdditionOptions>(
  forgeType: 'bitbucket-dc',
  key: T,
): BitbucketAdditionOptions[T];
// eslint-disable-next-line no-redeclare
function getAdditionalOptions<T extends keyof AddonAdditionOptions>(
  forgeType: 'addon',
  key: T,
): AddonAdditionOptions[T];
// eslint-disable-next-line no-redeclare
function getAdditionalOptions<T extends keyof Record<string, unknown>>(_forgeType: ForgeType, key: T): unknown {
  return forge.value?.additional_options?.[key];
}

function setAdditionalOptions<T extends keyof GitHubAdditionOptions>(
  forgeType: 'github',
  key: T,
  value: GitHubAdditionOptions[T],
): void;
// eslint-disable-next-line no-redeclare
function setAdditionalOptions<T extends keyof BitbucketAdditionOptions>(
  forgeType: 'bitbucket-dc',
  key: T,
  value: BitbucketAdditionOptions[T],
): void;
// eslint-disable-next-line no-redeclare
function setAdditionalOptions<T extends keyof AddonAdditionOptions>(
  forgeType: 'addon',
  key: T,
  value: AddonAdditionOptions[T],
): void;
// eslint-disable-next-line no-redeclare
function setAdditionalOptions<T extends keyof Record<string, unknown>>(
  _forgeType: ForgeType,
  key: string,
  value: T,
): void {
  forge.value = {
    ...forge.value,
    additional_options: {
      ...forge.value?.additional_options,
      [key]: value,
    },
  };
}

const replaceRegex = /\/$/;

const oauthAppForgeUrl = computed(() => {
  if (!forge.value || !forge.value.type || !forge.value.url) {
    return '';
  }

  const forgeUrl = `${forge.value.url.startsWith('http') ? '' : 'https://'}${forge.value.url.replace(replaceRegex, '')}`;

  switch (forge.value.type) {
    case 'github':
      return `${forgeUrl}/settings/applications/new`;
    case 'gitlab':
      return `${forgeUrl}/-/user_settings/applications`;
    case 'gitea':
    case 'forgejo':
      return `${forgeUrl}/user/settings/applications`;
    case 'bitbucket':
    case 'bitbucket-dc':
      return `${forgeUrl}/account/settings/app-passwords`;
    default:
      return '';
  }
});

const githubAppSettingsUrl = computed(() => {
  if (forge.value?.type !== 'github' || !forge.value.url) {
    return '';
  }

  const forgeUrl = `${forge.value.url.startsWith('http') ? '' : 'https://'}${forge.value.url.replace(replaceRegex, '')}`;
  return `${forgeUrl}/settings/apps/new`;
});

const apiClient = useApiClient();
const appHealth = ref<ForgeAppHealth>();
const { doSubmit: checkGithubApp, isLoading: isCheckingApp } = useAsyncAction(async () => {
  appHealth.value = undefined;
  if (!forge.value.id) {
    return;
  }
  appHealth.value = await apiClient.getForgeAppHealth(forge.value.id);
});

// the private key is write-only: the server never returns it, but marks
// whether a key is stored
const hasStoredAppKey = computed(() => !!forge.value.additional_options?.['app-private-key-set']);

watch(
  () => [forge.value.id, getAdditionalOptions('github', 'app-id'), getAdditionalOptions('github', 'app-private-key')],
  () => {
    appHealth.value = undefined;
  },
);

const forgeType = computed({
  get: () => forge.value?.type ?? forgeTypeOptions[0].value,
  set: (value) => {
    const newUrl = getDefaultUrl(value as ForgeType);

    // Only update URL if it hasn't been customized or is empty
    if (!forge.value?.url || forge.value.url === getDefaultUrl(forge.value.type as ForgeType)) {
      forge.value = { ...forge.value, url: newUrl, type: value as ForgeType };
    } else {
      forge.value = { ...forge.value, type: value as ForgeType };
    }
  },
});

const redirectUri = computed(() => [window.location.origin, config.rootPath, 'authorize'].filter((a) => !!a).join('/'));

async function submit() {
  if (!forge.value.url?.startsWith('http')) {
    forge.value.url = `https://${forge.value.url}`;
  }

  if (forge.value.oauth_host === forge.value.url) {
    forge.value.oauth_host = '';
  }

  if (forge.value.oauth_host && !forge.value.oauth_host.startsWith('http')) {
    forge.value.oauth_host = `https://${forge.value.oauth_host}`;
  }

  emit('submit');
}
</script>
