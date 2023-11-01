<template>
  <div v-if="branches" class="space-y-4">
    <ListItem
      v-for="branch in branches"
      :key="branch"
      class="text-wp-text-100"
      :to="{ name: 'repo-branch', params: { branch } }"
    >
      {{ branch }}
      <Badge v-if="branch === repo?.default_branch" :label="$t('default')" class="ml-auto" />
    </ListItem>
  </div>
</template>

<script lang="ts" setup>
import { inject, Ref, watch } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import useApiClient from '~/compositions/useApiClient';
import { usePagination } from '~/compositions/usePaginate';
import { Repo } from '~/lib/api/types';

const apiClient = useApiClient();

const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Unexpected: "repo" should be provided at this place');
}

async function loadBranches(page: number): Promise<string[]> {
  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  return apiClient.getRepoBranches(repo.value.id, page);
}

const { resetPage, data: branches } = usePagination(loadBranches);

watch(repo, resetPage);
</script>
