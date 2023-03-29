<template>
  <div v-if="branches" class="space-y-4">
    <ListItem
      v-for="branch in branches"
      :key="branch"
      class="text-color"
      :to="{ name: 'repo-branch', params: { branch } }"
    >
      {{ branch }}
    </ListItem>
  </div>
</template>

<script lang="ts" setup>
import { inject, onMounted, onUnmounted, Ref, ref, watch } from 'vue';

import ListItem from '~/components/atomic/ListItem.vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import { PaginatedList } from '~/compositions/usePaginate';

const apiClient = useApiClient();

const branches = ref<string[]>();
const repo = inject<Ref<Repo>>('repo');
if (!repo) {
  throw new Error('Unexpected: "repo" and "scrollComponent" should be provided at this place');
}

const list = new PaginatedList(loadBranches);

// TODO it seems this also runs if Pr list is open
async function loadBranches(page: number): Promise<boolean> {
  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  const _branches = await apiClient.getRepoBranches(repo.value.owner, repo.value.name, page);

  if (page === 1) {
    branches.value = _branches;
  } else {
    branches.value?.push(..._branches);
  }
  return _branches.length !== 0;
}

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
