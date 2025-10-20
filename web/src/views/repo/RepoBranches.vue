<template>
  <div class="space-y-4">
    <ListItem
      v-for="branch in branchesWithDefaultBranchFirst"
      :key="branch"
      class="text-wp-text-100"
      :to="{ name: 'repo-branch', params: { branch } }"
    >
      {{ branch }}
      <Badge v-if="branch === repo?.default_branch" :value="$t('default')" class="ml-auto" />
    </ListItem>
    <div v-if="loading" class="text-wp-text-100 flex justify-center">
      <Icon name="spinner" />
    </div>
    <Panel v-else-if="branches.length === 0" class="flex justify-center">
      {{ $t('empty_list', { entity: $t('repo.branches') }) }}
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';

const apiClient = useApiClient();

const repo = requiredInject('repo');

async function loadBranches(page: number): Promise<string[]> {
  return apiClient.getRepoBranches(repo.value.id, { page });
}

const { resetPage, data: branches, loading } = usePagination(loadBranches);

const branchesWithDefaultBranchFirst = computed(() =>
  branches.value.toSorted((a, b) => {
    if (a === repo.value.default_branch) {
      return -1;
    }

    if (b === repo.value.default_branch) {
      return 1;
    }

    return 0;
  }),
);

watch(repo, resetPage);

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.branches'), repo.value.full_name]));
</script>
