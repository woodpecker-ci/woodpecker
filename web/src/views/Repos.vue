<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('repositories.title') }}
    </template>

    <template #headerActions>
      <Button :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </template>

    <Transition name="fade" mode="out-in">
      <div v-if="search === '' && repos.length > 0" class="grid gap-8">
        <div v-if="reposLastAccess.length > 0 && repos.length > 4" class="flex flex-col gap-4">
          <div>
            <h2 class="text-wp-text-100 text-lg">{{ $t('repositories.last.title') }}</h2>
            <span class="text-wp-text-alt-100">{{ $t('repositories.last.desc') }}</span>
          </div>
          <div class="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-2">
            <RepoItem v-for="repo in reposLastAccess" :key="repo.id" :repo="repo" />
          </div>
        </div>

        <div class="flex flex-col gap-4">
          <div>
            <h2 class="text-wp-text-100 text-lg">{{ $t('repositories.all.title') }}</h2>
            <span class="text-wp-text-alt-100">{{ $t('repositories.all.desc') }}</span>
          </div>
          <div class="flex flex-col gap-4">
            <RepoItem v-for="repo in reposLastActivity" :key="repo.id" :repo="repo" />
          </div>
        </div>
      </div>
      <div v-else class="flex flex-col">
        <div v-if="reposLastActivity.length > 0" class="flex flex-col gap-4">
          <RepoItem v-for="repo in reposLastActivity" :key="repo.id" :repo="repo" />
        </div>
        <span v-else class="text-wp-text-100 text-center text-lg">{{ $t('no_search_results') }}</span>
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

const reposLastAccess = computed(() => sortReposByLastAccess(repos.value || []).slice(0, 4));

const search = ref('');
const { searchedRepos } = useRepoSearch(repos, search);
const reposLastActivity = computed(() => sortReposByLastActivity(searchedRepos.value || []));

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
