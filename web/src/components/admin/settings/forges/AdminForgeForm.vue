<template>
  <form @submit.prevent="submit">
    <Warning v-if="!isNew && forge.id === 1" :text="$t('forge_managed_by_env')" />

    <InputField v-slot="{ id }" :label="$t('forge_type')">
      <SelectField
        :id="id" v-model="forgeType" :options="[
          { value: 'github', text: $t('github') },
          { value: 'gitlab', text: $t('gitlab') },
          { value: 'gitea', text: $t('gitea') },
          { value: 'bitbucket', text: $t('bitbucket') },
          { value: 'forgejo', text: $t('forgejo') },
          { value: 'addon', text: $t('addon') },
        ]"
        required
      />
    </InputField>

    <InputField v-slot="{ id }" :label="$t('url')">
      <TextField :id="id" v-model="forge.url" required />
    </InputField>

    <hr class="my-4 border-gray-800" />

    <template v-if="forge.type && forge.url">
      <InputField v-slot="{ id }" :label="$t('oauth_redirect_uri')">
        <i18n-t keypath="use_this_redirect_uri_to_create" tag="p" class="mb-2">
          <a rel="noopener noreferrer" :href="oauthAppForgeUrl" target="_blank" class="underline">{{  $t('developer_settings') }}</a>
        </i18n-t>
        <TextField :id="id" :model-value="redirectUri" disabled />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('oauth_client_id')">
        <TextField :id="id" v-model="forge.client" required />
      </InputField>

      <InputField v-slot="{ id }" :label="$t('oauth_client_secret')">
        <TextField :id="id" v-model="forge.client_secret" :placeholder="isNew ? '' : $t('leave_empty_to_keep_current_value')" :required="isNew" />
      </InputField>

      <Panel
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
        </template>
        <template v-if="forge.type === 'bitbucket-dc'">
          <InputField v-slot="{ id }" :label="$t('git_username')">
            <p>{{ $t('git_username_desc') }}</p>
            <TextField
              :id="id" :model-value="getAdditionalOptions('bitbucket-dc', 'git-username')"
              @update:model-value="setAdditionalOptions('bitbucket-dc', 'git-username', $event)"
            />
          </InputField>
          <InputField v-slot="{ id }" :label="$t('git_password')">
            <p>{{ $t('git_password_desc') }}</p>
            <TextField
              :id="id" :model-value="getAdditionalOptions('bitbucket-dc', 'git-password')"
              @update:model-value="setAdditionalOptions('bitbucket-dc', 'git-password', $event)"
            />
          </InputField>
        </template>
        <template v-if="forge.type === 'addon'">
          <InputField v-slot="{ id }" :label="$t('executable')">
            <TextField :id="id" :model-value="getAdditionalOptions('addon', 'executable')" @update:model-value="setAdditionalOptions('addon', 'executable', $event)" />
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
        <Button :text="$t('cancel')" @click="forge = {}" />

        <Button
          :is-loading="isSaving"
          type="submit"
          color="green"
          :text="isNew ? $t('add') : $t('save')"
        />
      </div>
    </template>
  </form>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import Button from '~/components/atomic/Button.vue';
import Warning from '~/components/atomic/Warning.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useConfig from '~/compositions/useConfig';
import type { Forge, ForgeType } from '~/lib/api/types';

defineProps<{
  isNew?: boolean;
  isSaving?: boolean;
}>();

const emit = defineEmits<{
  (e: 'submit'): void;
}>();

const config = useConfig();

const forge = defineModel<Partial<Forge>>('forge', {
  required: true,
});

interface GitHubAdditionOptions {
  'merge-ref'?: boolean;
  'public-only'?: boolean;
}

interface BitbucketAdditionOptions {
  'git-username'?: string;
  'git-password'?: string;
}

interface AddonAdditionOptions {
  executable?: string;
}

function getAdditionalOptions<T extends keyof GitHubAdditionOptions>(forgeType: 'github', key: T): GitHubAdditionOptions[T];
// eslint-disable-next-line no-redeclare
function getAdditionalOptions<T extends keyof BitbucketAdditionOptions>(forgeType: 'bitbucket-dc',  key: T): BitbucketAdditionOptions[T];
// eslint-disable-next-line no-redeclare
function getAdditionalOptions<T extends keyof AddonAdditionOptions>(forgeType: 'addon',  key: T): AddonAdditionOptions[T];
// eslint-disable-next-line no-redeclare
function getAdditionalOptions<T extends keyof Record<string, unknown>>(_forgeType: ForgeType, key: T):  unknown {
  return forge.value?.additional_options?.[key];
}

function setAdditionalOptions<T extends keyof GitHubAdditionOptions>(forgeType: 'github', key: T, value: GitHubAdditionOptions[T]): void;
// eslint-disable-next-line no-redeclare
function setAdditionalOptions<T extends keyof BitbucketAdditionOptions>(forgeType: 'bitbucket-dc', key: T, value: BitbucketAdditionOptions[T]): void;
// eslint-disable-next-line no-redeclare
function setAdditionalOptions<T extends keyof AddonAdditionOptions>(forgeType: 'addon',  key: T, value: AddonAdditionOptions[T]): void;
// eslint-disable-next-line no-redeclare
function setAdditionalOptions<T extends keyof  Record<string, unknown>>(_forgeType: ForgeType, key: string, value: T): void {
  forge.value = {
    ...forge.value,
    additional_options: {
      ...forge.value?.additional_options,
      [key]: value,
    },
  };
}

const oauthAppForgeUrl = computed(() => {
  if (!forge.value || !forge.value.type || !forge.value.url) {
    return '';
  }

  const forgeUrl = `${forge.value.url.startsWith('http') ? '' : 'https://'}${forge.value.url.replace(/\/$/, '')}`;

  switch (forge.value.type) {
    case 'github':
      return `${forgeUrl}/settings/applications/new`;
    case 'gitlab':
      return `${forgeUrl}/-/user_settings/applications`;
    case 'gitea':
    case 'forgejo':
      return `${forgeUrl}/user/settings/applications`;
    case 'bitbucket':
      return `${forgeUrl}/account/settings/app-passwords`;
    case 'bitbucket-dc':
      return `${forgeUrl}/account/settings/app-passwords`;
    case 'addon':
      return '';
  }
});

const forgeType = computed({
  get: () => forge.value?.type ?? '',
  set: (value) => {
    let url = forge.value?.url || '';

    if (value === 'github') {
      url = url || 'github.com';
    } else if (value === 'gitlab') {
      url = url || 'gitlab.com';
    } else if (value === 'bitbucket') {
      url = url || 'bitbucket.org';
    }

    forge.value = { ...forge.value, url, type: value as ForgeType };
  },
});

const redirectUri = computed(() => [window.location.origin, config.rootPath, 'callback'].filter(a => !!a).join('/'));

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
