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
import { inject, onMounted, Ref, ref, watch } from 'vue';

import ListItem from '~/components/atomic/ListItem.vue';
import useApiClient from '~/compositions/useApiClient';
import { PullRequest, Repo } from '~/lib/api/types';

const apiClient = useApiClient();

const pullRequests = ref<PullRequest[]>();
const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Unexpected: "repo" should be provided at this place');
}

async function loadPullRequests() {
  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  pullRequests.value = await apiClient.getRepoPullRequests(repo.value.owner, repo.value.name);
}

onMounted(() => {
  loadPullRequests();
});

watch(repo, () => {
  loadPullRequests();
});
</script>
