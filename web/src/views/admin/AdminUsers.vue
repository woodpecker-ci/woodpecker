<template>
  <Settings :title="$t('admin.settings.users.users')" :description="$t('admin.settings.users.desc')">
    <template #headerActions>
      <Button
        v-if="selectedUser"
        :text="$t('admin.settings.users.show')"
        start-icon="back"
        @click="selectedUser = undefined"
      />
      <Button v-else :text="$t('admin.settings.users.add')" start-icon="plus" @click="showAddUser" />
    </template>

    <div v-if="!selectedUser" class="text-wp-text-100 space-y-4">
      <ListItem
        v-for="user in users"
        :key="user.id"
        class="bg-wp-background-200! dark:bg-wp-background-200! items-center gap-2"
      >
        <img v-if="user.avatar_url" class="h-6 rounded-md" :src="user.avatar_url" />
        <span>{{ user.login }}</span>
        <span class="ml-auto flex gap-2">
          <Badge
            v-if="forgesMap.has(user.forge_id)"
            class="md:display-unset hidden"
            :value="forgesMap.get(user.forge_id)"
          />
          <Badge v-if="user.admin" class="md:display-unset hidden" :value="$t('admin.settings.users.admin.admin')" />
        </span>
        <div class="flex items-center gap-2">
          <IconButton
            icon="edit"
            :title="$t('admin.settings.users.edit_user')"
            class="md:display-unset h-8 w-8"
            @click="editUser(user)"
          />
          <IconButton
            icon="trash"
            :title="$t('admin.settings.users.delete_user')"
            class="hover:text-wp-error-100 h-8 w-8"
            :is-loading="isDeleting"
            @click="deleteUser(user)"
          />
        </div>
      </ListItem>

      <div v-if="loading" class="flex justify-center">
        <Icon name="spinner" class="animate-spin" />
      </div>
      <div v-else-if="users?.length === 0" class="ml-2">{{ $t('admin.settings.users.none') }}</div>
    </div>
    <div v-else>
      <form @submit.prevent="saveUser">
        <InputField v-slot="{ id }" :label="$t('admin.settings.users.login')">
          <TextField
            :id="id"
            v-model="selectedUser.login"
            :placeholder="$t('admin.settings.users.login')"
            :disabled="isEditingUser"
          />
        </InputField>

        <InputField
          v-if="selectedUser!.forge_id !== undefined && forgesMap.has(selectedUser!.forge_id)"
          v-slot="{ id }"
          :label="$t('admin.settings.users.forge')"
        >
          <TextField :id="id" v-model="selectedUserForge" :placeholder="$t('admin.settings.users.forge')" disabled />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('admin.settings.users.email')">
          <TextField :id="id" v-model="selectedUser.email" :placeholder="$t('admin.settings.users.email')" />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('admin.settings.users.avatar_url')">
          <div class="flex gap-2">
            <img v-if="selectedUser.avatar_url" class="h-8 w-8 rounded-md" :src="selectedUser.avatar_url" />
            <TextField
              :id="id"
              v-model="selectedUser.avatar_url"
              login
              :placeholder="$t('admin.settings.users.avatar_url')"
            />
          </div>
        </InputField>

        <InputField :label="$t('admin.settings.users.admin.admin')">
          <Warning
            v-if="selectedUser.admin_env"
            class="mb-4 text-sm"
            :text="$t('admin.settings.users.admin.admin_warning')"
          />

          <Checkbox
            :model-value="selectedUser.admin || false"
            :label="$t('admin.settings.users.admin.placeholder')"
            @update:model-value="selectedUser!.admin = $event"
          />
        </InputField>

        <div class="flex gap-2">
          <Button :text="$t('admin.settings.users.cancel')" @click="selectedUser = undefined" />

          <Button
            :is-loading="isSaving"
            type="submit"
            color="green"
            :text="isEditingUser ? $t('admin.settings.users.save') : $t('admin.settings.users.add')"
          />
        </div>
      </form>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Warning from '~/components/atomic/Warning.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { User } from '~/lib/api/types';
import { deepClone } from '~/lib/utils';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

const forgesMap = ref<Map<number, string>>(new Map());

const selectedUser = ref<Partial<User>>();
const isEditingUser = computed(() => !!selectedUser.value?.id);
const selectedUserForge = computed(() => forgesMap.value.get(selectedUser.value?.forge_id || -1));

async function loadForges() {
  const forges = await apiClient.getForges({ page: 1 });
  if (forges) {
    forgesMap.value = new Map(
      forges.map((forge) => {
        let name = forge.type.charAt(0).toUpperCase() + forge.type.slice(1);

        if (forge.url || forge.oauth_host) {
          const url = new URL(forge.oauth_host || forge.url);
          name = url.hostname;
        }

        return [forge.id, name];
      }),
    );
  }
}

onMounted(loadForges);

async function loadUsers(page: number): Promise<User[] | null> {
  return apiClient.getUsers({ page });
}

const { resetPage, data: users, loading } = usePagination(loadUsers, () => !selectedUser.value);

const { doSubmit: saveUser, isLoading: isSaving } = useAsyncAction(async () => {
  if (!selectedUser.value) {
    throw new Error("Unexpected: Can't get user");
  }

  if (isEditingUser.value) {
    await apiClient.updateUser(selectedUser.value);
    notifications.notify({
      title: t('admin.settings.users.saved'),
      type: 'success',
    });
  } else {
    selectedUser.value = await apiClient.createUser(selectedUser.value);
    notifications.notify({
      title: t('admin.settings.users.created'),
      type: 'success',
    });
  }
  selectedUser.value = undefined;
  await resetPage();
});

const { doSubmit: deleteUser, isLoading: isDeleting } = useAsyncAction(async (_user: User) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('admin.settings.users.delete_confirm'))) {
    return;
  }

  await apiClient.deleteUser(_user);
  notifications.notify({ title: t('admin.settings.users.deleted'), type: 'success' });
  await resetPage();
});

function editUser(user: User) {
  selectedUser.value = deepClone(user);
}

function showAddUser() {
  selectedUser.value = { login: '' };
}

useWPTitle(computed(() => [t('admin.settings.users.users'), t('admin.settings.settings')]));
</script>
