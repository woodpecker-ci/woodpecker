<template>
  <div v-if="pullRequests" class="space-y-4">
    <ListItem
      v-for="pullRequest in pullRequests"
      :key="pullRequest.index"
      class="text-wp-text-100"
      :to="{ name: 'repo-pull-request', params: { pullRequest: pullRequest.index } }"
    >
      <span class="text-wp-text-alt-100 <md:hidden">#{{ pullRequest.index }}</span>
      <span class="text-wp-text-alt-100 <md:hidden mx-2">-</span>
      <span class="text-wp-text-100 <md:underline whitespace-nowrap overflow-hidden overflow-ellipsis">{{
        pullRequest.title
      }}</span>
    </ListItem>
  </div>
</template>

<script lang="ts" setup>
import { inject, Ref, watch } from 'vue';

import ListItem from '~/components/atomic/ListItem.vue';
import useApiClient from '~/compositions/useApiClient';
import { usePagination } from '~/compositions/usePaginate';
import { PullRequest, Repo } from '~/lib/api/types';

const apiClient = useApiClient();

const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Unexpected: "repo" should be provided at this place');
}
if (!repo.value.pr_enabled || !repo.value.allow_pr) {
  throw new Error('Unexpected: pull requests are disabled for repo');
}

async function loadPullRequests(page: number): Promise<PullRequest[]> {
  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  return apiClient.getRepoPullRequests(repo.value.id, page);
}

const { resetPage, data: pullRequests } = usePagination(loadPullRequests);

watch(repo, resetPage);
</script>
