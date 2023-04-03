<template>
  <div v-if="pullRequests" class="space-y-4">
    <ListItem
      v-for="pullRequest in pullRequests"
      :key="pullRequest.index"
      class="text-color"
      :to="{ name: 'repo-pull-request', params: { pullRequest: pullRequest.index } }"
    >
      <span class="text-color-alt <md:hidden">#{{ pullRequest.index }}</span>
      <span class="text-color-alt <md:hidden mx-2">-</span>
      <span class="text-color <md:underline whitespace-nowrap overflow-hidden overflow-ellipsis">{{
        pullRequest.title
      }}</span>
    </ListItem>
  </div>
</template>

<script lang="ts" setup>
import {inject, onBeforeUnmount, onMounted, onUnmounted, Ref, ref, watch} from 'vue';

import ListItem from '~/components/atomic/ListItem.vue';
import useApiClient from '~/compositions/useApiClient';
import { PaginatedList } from '~/compositions/usePaginate';
import { PullRequest, Repo } from '~/lib/api/types';

const apiClient = useApiClient();

const pullRequests = ref<PullRequest[]>();
const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Unexpected: "repo" should be provided at this place');
}

async function loadPullRequests(page: number): Promise<boolean> {
  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  const pulls = await apiClient.getRepoPullRequests(repo.value.owner, repo.value.name, page);

  if (page === 1) {
    pullRequests.value = pulls;
  } else {
    pullRequests.value?.push(...pulls);
  }
  return pulls.length !== 0;
}

const list = new PaginatedList(loadPullRequests);

onMounted(() => {
  list.onMounted();
});

onUnmounted(() => {
  list.onUnmounted();
});

watch(repo, () => {
  list.reset(true);
});
</script>
