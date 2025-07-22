<template>
  <Settings :title="$t('admin.settings.repos.repos')" :description="$t('admin.settings.repos.desc')">
    <template #headerActions>
      <Button
        start-icon="heal"
        :is-loading="isRepairingRepos"
        :text="$t('admin.settings.repos.repair.repair')"
        @click="repairRepos"
      />
    </template>

    <div class="text-wp-text-100 space-y-4">
      <ListItem
        v-for="repo in repos"
        :key="repo.id"
        class="bg-wp-background-200 dark:bg-wp-background-100! items-center gap-2"
      >
        <span>{{ repo.full_name }}</span>
        <div class="ml-auto flex items-center">
          <Badge
            v-if="!repo.active"
            class="md:display-unset mr-2 hidden"
            :value="$t('admin.settings.repos.disabled')"
          />
          <IconButton
            icon="chevron-right"
            :title="$t('admin.settings.repos.view')"
            class="h-8 w-8"
            :to="{ name: 'repo', params: { repoId: repo.id } }"
          />
          <IconButton
            icon="settings-outline"
            :title="$t('admin.settings.repos.settings')"
            class="h-8 w-8"
            :to="{ name: 'repo-settings', params: { repoId: repo.id } }"
          />
        </div>
      </ListItem>

      <div v-if="loading" class="flex justify-center">
        <Icon name="spinner" class="animate-spin" />
      </div>
      <div v-else-if="repos?.length === 0" class="ml-2">{{ $t('admin.settings.repos.none') }}</div>
    </div>
  </Settings>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
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
import { useWPTitle } from '~/compositions/useWPTitle';
import type { Repo } from '~/lib/api/types';

const apiClient = useApiClient();
const notifications = useNotifications();
const i18n = useI18n();

async function loadRepos(page: number): Promise<Repo[] | null> {
  return apiClient.getAllRepos({ page });
}

const { data: repos, loading } = usePagination(loadRepos);

const { doSubmit: repairRepos, isLoading: isRepairingRepos } = useAsyncAction(async () => {
  await apiClient.repairAllRepos();
  notifications.notify({ title: i18n.t('admin.settings.repos.repair.success'), type: 'success' });
});

useWPTitle(computed(() => [i18n.t('admin.settings.repos.repos'), i18n.t('admin.settings.settings')]));
</script>
