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

const apiClient = useApiClient();

let page = 1;
let getNextPage = true;

const branches = ref<string[]>();
const repo = inject<Ref<Repo>>('repo');
const scrollComponent = document.querySelector('main > div');
if (!repo || !scrollComponent) {
  throw new Error('Unexpected: "repo" and "scrollComponent" should be provided at this place');
}

async function loadBranches() {
  getNextPage = false;
  if (!repo) {
    throw new Error('Unexpected: "repo" should be provided at this place');
  }

  const _branches = await apiClient.getRepoBranches(repo.value.owner, repo.value.name, page);

  if (page === 1) {
    branches.value = _branches;
  } else {
    branches.value?.push(..._branches);
  }
  getNextPage = _branches.length !== 0;
}

const handleScroll = () => {
  if (getNextPage && scrollComponent.scrollTop + scrollComponent.clientHeight === scrollComponent.scrollHeight) {
    page += 1;
    loadBranches();
  }
};

onMounted(() => {
  page = 1;
  loadBranches();
  scrollComponent.addEventListener('scroll', handleScroll);
});

onUnmounted(() => {
  page = 1;
  scrollComponent.removeEventListener('scroll', handleScroll);
});

watch(repo, () => {
  page = 1;
  loadBranches();
});
</script>
