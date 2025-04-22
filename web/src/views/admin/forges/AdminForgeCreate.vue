<template>
  <Settings :title="$t('forges')" :description="$t('forges_desc')">
    <template #headerActions>
      <Button
        :text="$t('show_forges')"
        start-icon="back"
        :to="{ name: 'admin-settings-forges' }"
      />
    </template>

    <form @submit.prevent="saveForge">
      <template v-if="step === 1">
        <InputField v-slot="{ id }" :label="$t('forge_type')">
          <SelectField
            :id="id" :model-value="forge.type || ''" :options="[
              { value: 'github', text: $t('github') },
              { value: 'gitlab', text: $t('gitlab') },
              { value: 'gitea', text: $t('gitea') },
              { value: 'bitbucket', text: $t('bitbucket') },
              { value: 'forgejo', text: $t('forgejo') },
              { value: 'addon', text: $t('addon') },
            ]" @update:model-value="forge.type = $event as ForgeType"
          />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('url')">
          <TextField :id="id" v-model="forge.url" />
        </InputField>
      </template>

      <template v-if="step === 2">
        <p>Please create an OAuth app at ... and paste the credentials you've received here:</p>

        <InputField v-slot="{ id }" :label="$t('oauth_client_id')">
          <TextField :id="id" v-model="forge.client" />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('oauth_client_secret')">
          <TextField :id="id" v-model="forge.client_secret" />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('oauth_host')">
          <TextField :id="id" v-model="forge.oauth_host" />
        </InputField>
      </template>

      <template v-if="step === 3">
        <InputField :label="$t('skip_verify')">
          <Checkbox
            :label="$t('skip_verify_desc')"
            :model-value="forge.skip_verify || false"
            @update:model-value="forge!.skip_verify = $event"
          />
        </InputField>


        <template v-if="forge.type === 'github'">
          <InputField :label="$t('merge_ref')">
            <Checkbox
              :label="$t('merge_ref_desc')"
              :model-value="getAdditionalOptions('github', 'merge-ref')"
              @update:model-value="setAdditionOptions('merge-ref', $event)"
            />
          </InputField>

          <InputField :label="$t('public_only')">
            <Checkbox
              :label="$t('public_only_desc')"
              :model-value="forge.additional_options?.['public-only'] ?? false"
              @update:model-value="setAdditionOptions('public-only', $event)"
            />
          </InputField>
        </template>
        <template v-if="forge.type === 'bitbucket-dc'">
          <InputField v-slot="{ id }" :label="$t('git_username')">
            <p>{{ $t('git_username_desc') }}</p>
            <TextField :id="id" :model-value="getAdditionOptions('bitbucket-dc', 'git-username')" />
          </InputField>
          <InputField v-slot="{ id }" :label="$t('git_password')">
            <p>{{ $t('git_password_desc') }}</p>
            <TextField :id="id" :model-value="getAdditionOptions('bitbucket-dc', 'git-password')" />
          </InputField>
        </template>
        <template v-if="forge.type === 'addon'">
          <InputField v-slot="{ id }" :label="$t('executable')">
            <TextField :id="id" :model-value="getAdditionOptions('addon', 'executable')" />
          </InputField>
        </template>
      </template>


      <div class="flex gap-2">
        <Button v-if="step > 1" type="button" :text="$t('previous')" @click="step--" />
        <Button v-if="step < 3" type="button" :text="$t('next')" @click="step++" />

        <template v-if="step === 3">
          <Button :text="$t('cancel')" @click="forge = {}" />

          <Button
            :is-loading="isSaving"
            type="submit"
            color="green"
            :text="$t('add')"
          />
        </template>
      </div>
    </form>
  </Settings>
</template>

<script lang="ts" setup>
import { ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Button from '~/components/atomic/Button.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import SelectField from '~/components/form/SelectField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import type { Forge, ForgeType } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

const forge = ref<Partial<Forge>>({});

const step = ref(1);

const { doSubmit: saveForge, isLoading: isSaving } = useAsyncAction(async () => {
  if (!forge.value) {
    throw new Error("Unexpected: Can't get forge");
  }

  forge.value = await apiClient.createForge(forge.value);
  notifications.notify({
    title: t('admin.settings.users.created'),
    type: 'success',
  });

  // TODO: redirect to forge edit
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

function getAdditionalOptions<T extends keyof GitHubAdditionOptions>(forgeType: 'github', forge: Forge, key: T): GitHubAdditionOptions[T];
function getAdditionalOptions<T extends keyof Record<string, unknown>>(forgeType: ForgeType, forge: Forge, key: T):  unknown {
  if (forgeType === 'github') {
    if (key === 'merge-ref') {
      return forge.additional_options?.['merge-ref'] ?? false;
    }

    if (key === 'public-only') {
      return forge.additional_options?.['public-only'] ?? false;
    }
  }

  return undefined;
}

function setAdditionOptions<T extends keyof GitHubAdditionOptions>(key: T, value: GitHubAdditionOptions[T]): void;
function setAdditionOptions<T extends keyof  Record<string, unknown>>(key: string, value: unknown): void {
  if (!forge.value) {
    throw new Error("Unexpected: Can't get forge");
  }

  forge.value = {
    ...forge.value,
    additional_options: {
      ...forge.value.additional_options ?? {},
      [key]: value,
    },
  };
}
</script>
