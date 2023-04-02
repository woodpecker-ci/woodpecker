<template>
  <Panel>
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-gray-600">
      <div class="ml-2">
        <h1 class="text-xl text-color">{{ $t('admin.settings.users.users') }}</h1>
        <p class="text-sm text-color-alt">{{ $t('admin.settings.users.desc') }}</p>
      </div>
      <Button
        v-if="selectedUser"
        class="ml-auto"
        :text="$t('admin.settings.users.show')"
        start-icon="back"
        @click="selectedUser = undefined"
      />
      <Button v-else class="ml-auto" :text="$t('admin.settings.users.add')" start-icon="plus" @click="showAddUser" />
    </div>

    <div v-if="!selectedUser" class="space-y-4 text-color">
      <ListItem v-for="user in users" :key="user.id" class="items-center gap-2">
        <img v-if="user.avatar_url" class="rounded-md h-6" :src="user.avatar_url" />
        <span>{{ user.login }}</span>
        <Badge
          v-if="user.admin"
          class="ml-auto hidden md:inline-block"
          :label="$t('admin.settings.users.admin.admin')"
        />
        <IconButton
          icon="edit"
          :title="$t('admin.settings.users.edit_user')"
          class="w-8 h-8"
          :class="{ 'ml-auto': !user.admin, 'ml-2': user.admin }"
          @click="editUser(user)"
        />
        <IconButton
          icon="trash"
          :title="$t('admin.settings.users.delete_user')"
          class="ml-2 w-8 h-8 hover:text-red-400 hover:dark:text-red-500"
          :is-loading="isDeleting"
          @click="deleteUser(user)"
        />
      </ListItem>

      <div v-if="users?.length === 0" class="ml-2">{{ $t('admin.settings.users.none') }}</div>
    </div>
    <div v-else>
      <form @submit.prevent="saveUser">
        <InputField :label="$t('admin.settings.users.login')">
          <TextField v-model="selectedUser.login" :disabled="isEditingUser" />
        </InputField>

        <InputField :label="$t('admin.settings.users.email')">
          <TextField v-model="selectedUser.email" />
        </InputField>

        <InputField :label="$t('admin.settings.users.avatar_url')">
          <div class="flex gap-2">
            <img v-if="selectedUser.avatar_url" class="rounded-md h-8 w-8" :src="selectedUser.avatar_url" />
            <TextField v-model="selectedUser.avatar_url" />
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
  </Panel>
</template>

<script lang="ts" setup>
import { cloneDeep } from 'lodash';
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import InputField from '~/components/form/InputField.vue';
import TextField from '~/components/form/TextField.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { PaginatedList } from '~/compositions/usePaginate';
import { User } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

const users = ref<User[]>([]);
const selectedUser = ref<Partial<User>>();
const isEditingUser = computed(() => !!selectedUser.value?.id);
const list = new PaginatedList(loadUsers);

async function loadUsers(page: number): Promise<boolean> {
  const u = await apiClient.getUsers(page);
  if (page === 1 && u !== null) {
    users.value = u;
  } else if (u != null) {
    users.value?.push(...u);
  }
  return u != null && u.length != 0;
}

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
    selectedUser.value = undefined;
  } else {
    selectedUser.value = await apiClient.createUser(selectedUser.value);
    notifications.notify({
      title: t('admin.settings.users.created'),
      type: 'success',
    });
  }
  list.reset(true);
});

const { doSubmit: deleteUser, isLoading: isDeleting } = useAsyncAction(async (_user: User) => {
  // eslint-disable-next-line no-restricted-globals, no-alert
  if (!confirm(t('admin.settings.users.delete_confirm'))) {
    return;
  }

  await apiClient.deleteUser(_user);
  notifications.notify({ title: t('admin.settings.users.deleted'), type: 'success' });
  list.reset(true);
});

function editUser(user: User) {
  selectedUser.value = cloneDeep(user);
}

function showAddUser() {
  selectedUser.value = cloneDeep({ login: '' });
}

onMounted(async () => {
  list.onMounted();
});

onUnmounted(() => {
  list.onUnmounted();
});
</script>
