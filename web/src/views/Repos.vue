<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('repositories') }}
    </template>

    <template #headerActions>
      <Button :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </template>

    <div class="flex flex-col gap-12">
      <div class="gap-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2">
        <RepoItems v-for="repo in repoListAccess" :key="repo.id" :repo="repo" />
      </div>

      <div class="flex flex-col gap-4">
        <RepoItems v-for="repo in repoListActivity" :key="repo.id" :repo="repo" />
      </div>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import RepoItems from '~/components/repo/RepoItems.vue';
import useRepos from '~/compositions/useRepos';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { useRepoStore } from '~/store/repos';

const repoStore = useRepoStore();
const repos = computed(() => Object.values(repoStore.ownedRepos));
const search = ref('');

const { searchedRepos } = useRepoSearch(repos, search);
const { sortReposByLastAccess, sortReposByLastActivity } = useRepos();

const repoListAccess = computed(() => sortReposByLastAccess(repos.value || []));
const repoListActivity = computed(() => sortReposByLastActivity(searchedRepos.value || []));

onMounted(async () => {
  await repoStore.loadRepos();
});
</script>
