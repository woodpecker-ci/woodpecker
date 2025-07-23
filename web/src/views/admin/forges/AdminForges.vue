<template>
  <Settings :title="$t('forges')" :description="$t('forges_desc')">
    <template #headerActions>
      <Button :text="$t('add_forge')" start-icon="plus" :to="{ name: 'admin-settings-forge-create' }" />
    </template>

    <div class="text-wp-text-100 space-y-4">
      <ListItem
        v-for="forge in forges"
        :key="forge.id"
        class="bg-wp-background-200! dark:bg-wp-background-100! items-center gap-2"
      >
        <span>{{ forge.url.replace(/http(s):\/\//, '') }}</span>
        <Badge class="md:display-unset ml-auto hidden" :value="forge.type" />
        <IconButton
          icon="edit"
          :title="$t('edit_forge')"
          class="md:display-unset h-8 w-8"
          :to="{ name: 'admin-settings-forge', params: { forgeId: forge.id } }"
        />
        <IconButton
          icon="trash"
          :title="$t('delete_forge')"
          class="hover:text-wp-error-100 ml-2 h-8 w-8"
          :is-loading="isDeleting"
          @click="deleteForge(forge)"
        />
      </ListItem>

      <div v-if="loading" class="flex justify-center">
        <Icon name="spinner" class="animate-spin" />
      </div>
      <div v-else-if="forges?.length === 0" class="ml-2">{{ $t('no_forges') }}</div>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import type { Forge } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

async function loadForges(page: number): Promise<Forge[] | null> {
  return apiClient.getForges({ page });
}

const { resetPage, data: forges, loading } = usePagination(loadForges);

const { doSubmit: deleteForge, isLoading: isDeleting } = useAsyncAction(async (_forge: Forge) => {
  // eslint-disable-next-line no-alert
  if (!confirm(t('forge_delete_confirm'))) {
    return;
  }

  await apiClient.deleteForge(_forge);
  notifications.notify({ title: t('forge_deleted'), type: 'success' });
  await resetPage();
});
</script>
