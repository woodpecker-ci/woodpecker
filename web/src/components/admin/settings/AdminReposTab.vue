<template>
  <Settings :title="$t('admin.settings.repos.repos')" :desc="$t('admin.settings.repos.desc')">
    <template #titleActions>
      <Button
        start-icon="heal"
        :is-loading="isRepairingRepos"
        :text="$t('admin.settings.repos.repair.repair')"
        @click="repairRepos"
      />
    </template>

    <div class="space-y-4 text-wp-text-100">
      <ListItem
        v-for="repo in repos"
        :key="repo.id"
        class="items-center gap-2 !bg-wp-background-200 !dark:bg-wp-background-100"
      >
        <span>{{ repo.full_name }}</span>
        <div class="ml-auto flex items-center">
          <Badge v-if="!repo.active" class="<md:hidden mr-2" :label="$t('admin.settings.repos.disabled')" />
          <IconButton
            icon="chevron-right"
            :title="$t('admin.settings.repos.view')"
            class="w-8 h-8"
            :to="{ name: 'repo', params: { repoId: repo.id } }"
          />
          <IconButton
            icon="settings"
            :title="$t('admin.settings.repos.settings')"
            class="w-8 h-8"
            :to="{ name: 'repo-settings', params: { repoId: repo.id } }"
          />
        </div>
      </ListItem>

      <div v-if="repos?.length === 0" class="ml-2">{{ $t('admin.settings.orgs.none') }}</div>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import Badge from '~/components/atomic/Badge.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Settings from '~/components/layout/Settings.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { usePagination } from '~/compositions/usePaginate';
import { Repo } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();

async function loadRepos(page: number): Promise<Repo[] | null> {
  return apiClient.getAllRepos(page);
}

const { data: repos } = usePagination(loadRepos);

const { doSubmit: repairRepos, isLoading: isRepairingRepos } = useAsyncAction(async () => {
  await apiClient.repairAllRepos();
  notifications.notify({ title: i18n.t('admin.settings.repos.repair.success'), type: 'success' });
});
</script>
