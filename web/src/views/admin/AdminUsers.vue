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
        class="bg-wp-background-200! dark:bg-wp-background-100! items-center gap-2"
      >
        <img v-if="user.avatar_url" class="h-6 rounded-md" :src="user.avatar_url" />
        <span>{{ user.login }}</span>
        <Badge
          v-if="user.admin"
          class="md:display-unset ml-auto hidden"
          :value="$t('admin.settings.users.admin.admin')"
        />
        <IconButton
          icon="edit"
          :title="$t('admin.settings.users.edit_user')"
          class="md:display-unset h-8 w-8"
          :class="{ 'ml-auto': !user.admin, 'ml-2': user.admin }"
          @click="editUser(user)"
        />
        <IconButton
          icon="trash"
          :title="$t('admin.settings.users.delete_user')"
          class="hover:text-wp-error-100 ml-2 h-8 w-8"
          :is-loading="isDeleting"
          @click="deleteUser(user)"
        />
      </ListItem>

      <div v-if="users?.length === 0" class="ml-2">{{ $t('admin.settings.users.none') }}</div>
    </div>
    <div v-else>
      <form @submit.prevent="saveUser">
        <InputField v-slot="{ id }" :label="$t('admin.settings.users.login')">
          <TextField :id="id" v-model="selectedUser.login" :disabled="isEditingUser" />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('admin.settings.users.email')">
          <TextField :id="id" v-model="selectedUser.email" />
        </InputField>

        <InputField v-slot="{ id }" :label="$t('admin.settings.users.avatar_url')">
          <div class="flex gap-2">
            <img v-if="selectedUser.avatar_url" class="h-8 w-8 rounded-md" :src="selectedUser.avatar_url" />
            <TextField :id="id" v-model="selectedUser.avatar_url" />
          </div>
        </InputField>

        <InputField :label="$t('admin.settings.users.admin.admin')">
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
import { cloneDeep } from 'lodash';
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
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

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

const selectedUser = ref<Partial<User>>();
const isEditingUser = computed(() => !!selectedUser.value?.id);

async function loadUsers(page: number): Promise<User[] | null> {
  return apiClient.getUsers({ page });
}

const { resetPage, data: users } = usePagination(loadUsers, () => !selectedUser.value);

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
  resetPage();
});

const { doSubmit: deleteUser, isLoading: isDeleting } = useAsyncAction(async (_user: User) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('admin.settings.users.delete_confirm'))) {
    return;
  }

  await apiClient.deleteUser(_user);
  notifications.notify({ title: t('admin.settings.users.deleted'), type: 'success' });
  resetPage();
});

function editUser(user: User) {
  selectedUser.value = cloneDeep(user);
}

function showAddUser() {
  selectedUser.value = cloneDeep({ login: '' });
}

useWPTitle(computed(() => [t('admin.settings.users.users'), t('admin.settings.settings')]));
</script>
