<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('repositories') }}
    </template>

    <template #titleActions>
      <Button :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </template>

    <div class="flex flex-col gap-12">
      <div class="gap-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2">
        <RepoListItem v-for="repo in repoListAccess" :key="repo.id" :repo="repo" />
      </div>

      <div class="flex flex-col gap-4">
        <RepoListItem v-for="repo in repoListActivit" :key="repo.id" :repo="repo" />
      </div>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import RepoListItem from '~/components/repo/RepoListItem.vue';
import useRepos from '~/compositions/useRepos';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { useRepoStore } from '~/store/repos';

const router = useRouter();

const repoStore = useRepoStore();
const repos = computed(() => Object.values(repoStore.ownedRepos));
const search = ref('');

const { searchedRepos } = useRepoSearch(repos, search);
const { sortReposByLastAccess, sortReposByLastActivity } = useRepos();

const repoListAccess = computed(() => sortReposByLastAccess(repos.value || []));
const repoListActivit = computed(() => sortReposByLastActivity(searchedRepos.value || []));

router.beforeEach(async () => {
  await repoStore.loadRepos();
});
</script>
