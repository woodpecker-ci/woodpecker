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
        <span class="truncate">{{ user.login }}</span>
        <div class="ml-auto flex items-center gap-2">
          <span :title="userForgeName(user)" class="flex items-center">
            <Icon :name="userForgeIcon(user)" class="text-wp-text-alt-100" />
          </span>
          <Badge class="md:display-unset hidden" :label="$t('forge_type')" :value="userForgeName(user)" />
          <Badge v-if="user.admin" class="md:display-unset hidden" :value="$t('admin.settings.users.admin.admin')" />
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
          <TextField :id="id" v-model="selectedUser.login" :disabled="isEditingUser" :placeholder="$t('username')" />
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

        <InputField v-if="isEditingUser" :label="$t('forge_type')">
          <div class="flex min-h-8 items-center gap-2">
            <Icon :name="userForgeIcon(selectedUser)" class="text-wp-text-alt-100" />
            <Badge :value="userForgeName(selectedUser)" />
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
import { computed, reactive, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import type { IconNames } from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Warning from '~/components/atomic/Warning.vue';
import Checkbox from '~/components/form/Checkbox.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useForgeStore } from '~/compositions/useForgeStore';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Forge, User } from '~/lib/api/types';
import { forgeDisplayName, forgeIconName } from '~/lib/forge-utils';
import { deepClone } from '~/lib/utils';

const apiClient = useApiClient();
const forgeStore = useForgeStore();
const notifications = useNotifications();
const { t } = useI18n();

const selectedUser = ref<Partial<User>>();
const isEditingUser = computed(() => !!selectedUser.value?.id);
const userForges = reactive(new Map<number, Forge>());

async function loadUsers(page: number): Promise<User[] | null> {
  return apiClient.getUsers({ page });
}

const { resetPage, data: users, loading } = usePagination(loadUsers, () => !selectedUser.value);

async function loadForge(forgeId?: number) {
  if (forgeId === undefined || userForges.has(forgeId)) {
    return;
  }

  const forge = (await forgeStore.getForge(forgeId)).value;
  if (forge) {
    userForges.set(forgeId, forge);
  }
}

function userForge(user?: Partial<User>): Forge | undefined {
  if (user?.forge_id === undefined) {
    return undefined;
  }

  return userForges.get(user.forge_id);
}

function userForgeName(user?: Partial<User>): string {
  return forgeDisplayName(userForge(user), user?.forge_id);
}

function userForgeIcon(user?: Partial<User>): IconNames {
  return forgeIconName(userForge(user));
}

watch(
  users,
  (loadedUsers) => {
    const forgeIds = new Set((loadedUsers ?? []).map((user) => user.forge_id));
    forgeIds.forEach((forgeId) => void loadForge(forgeId));
  },
  { immediate: true },
);

watch(selectedUser, (user) => void loadForge(user?.forge_id));

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
