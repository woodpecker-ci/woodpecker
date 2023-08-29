<template>
  <Settings :title="$t('admin.settings.orgs.orgs')" :desc="$t('admin.settings.orgs.desc')">
    <div class="space-y-4 text-wp-text-100">
      <ListItem
        v-for="org in orgs"
        :key="org.id"
        class="items-center gap-2 !bg-wp-background-200 !dark:bg-wp-background-100"
      >
        <span>{{ org.name }}</span>
        <IconButton
          icon="chevron-right"
          :title="$t('admin.settings.orgs.view')"
          class="ml-auto w-8 h-8"
          :to="{ name: 'org', params: { orgId: org.id } }"
        />
        <IconButton
          icon="settings"
          :title="$t('admin.settings.orgs.org_settings')"
          class="w-8 h-8"
          :to="{ name: 'org-settings', params: { orgId: org.id } }"
        />
        <IconButton
          icon="trash"
          :title="$t('admin.settings.orgs.delete_org')"
          class="ml-2 w-8 h-8 hover:text-wp-control-error-100"
          :is-loading="isDeleting"
          @click="deleteOrg(org)"
        />
      </ListItem>

      <div v-if="orgs?.length === 0" class="ml-2">{{ $t('admin.settings.orgs.none') }}</div>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { useI18n } from 'vue-i18n';

import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { Org } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const { t } = useI18n();

async function loadOrgs(page: number): Promise<Org[] | null> {
  return apiClient.getOrgs(page);
}

const { resetPage, data: orgs } = usePagination(loadOrgs);

const { doSubmit: deleteOrg, isLoading: isDeleting } = useAsyncAction(async (_org: Org) => {
  // eslint-disable-next-line no-restricted-globals, no-alert
  if (!confirm(t('admin.settings.orgs.delete_confirm'))) {
    return;
  }

  await apiClient.deleteOrg(_org);
  notifications.notify({ title: t('admin.settings.orgs.deleted'), type: 'success' });
  resetPage();
});
</script>
