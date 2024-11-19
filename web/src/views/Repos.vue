<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('repositories') }}
    </template>

    <template #headerActions>
      <Button :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </template>

    <Transition name="fade">
      <div v-if="search === ''" class="flex flex-col">
        <div class="gap-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2">
          <RepoItem v-for="repo in repoListLastAccess" :key="repo.id" :repo="repo" />
        </div>

        <p class="text-wp-text-100 mt-12 mb-2">{{ $t('all_repositories') }}</p>
        <div class="flex flex-col gap-4">
          <RepoItem v-for="repo in repoListActivity" :key="repo.id" :repo="repo" />
        </div>
      </div>

      <div v-else class="flex flex-col gap-4">
        <RepoItem v-for="repo in repoListActivity" :key="repo.id" :repo="repo" />
      </div>
    </Transition>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import RepoItem from '~/components/repo/RepoItem.vue';
import useRepos from '~/compositions/useRepos';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { useRepoStore } from '~/store/repos';

const repoStore = useRepoStore();

const { sortReposByLastAccess, sortReposByLastActivity, repoWithLastPipeline } = useRepos();
const repos = computed(() => Object.values(repoStore.ownedRepos).map((r) => repoWithLastPipeline(r)));

const repoListLastAccess = computed(() => sortReposByLastAccess(repos.value || []).slice(0, 4));

const search = ref('');
const { searchedRepos } = useRepoSearch(
  computed(() => {
    if (search.value === '') {
      return repos.value.filter((r) => !repoListLastAccess.value.includes(r));
    }

    return repos.value;
  }),
  search,
);
const repoListActivity = computed(() => sortReposByLastActivity(searchedRepos.value || []));

onMounted(async () => {
  await repoStore.loadRepos();
});
</script>

<style scoped>
.fade-enter-active,
.fade-leave-active {
  transition: all 0.2s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}
</style>
