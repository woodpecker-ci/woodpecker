<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('repositories') }}
    </template>

    <template #titleActions>
      <Button :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </template>

    <div class="gap-4 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-2">
      <router-link v-for="repo in repoList" :key="repo.id" :to="{ name: 'repo', params: { repoId: repo.id } }">
        <div
          class="flex flex-col border rounded-md bg-wp-background-100 overflow-hidden p-4 border-wp-background-400 dark:bg-wp-background-200 cursor-pointer hover:shadow-md hover:bg-wp-background-300 dark:hover:bg-wp-background-300"
        >
          <div class="flex items-center gap-2">
            <img v-if="repo.avatar_url" :src="repo.avatar_url" class="w-8 h-8" alt="..." />
            <Icon v-else name="repo" class="text-wp-text-100" />
            <span class="text-wp-text-100 text-xl">{{ `${repo.owner} / ${repo.name}` }}</span>
            <Badge v-if="repo.visibility === RepoVisibility.Public" label="public" class="ml-auto" />
          </div>

          <div class="mt-8 flex gap-2">
            <PipelineStatusIcon status="failure" />
            <span class="text-wp-text-100">last pipeline was successful 20 mins ago</span>
          </div>
        </div>
      </router-link>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import useRepos from '~/compositions/useRepos';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { RepoVisibility } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';

const repoStore = useRepoStore();
const repos = computed(() => Object.values(repoStore.ownedRepos));
const search = ref('');

const { searchedRepos } = useRepoSearch(repos, search);
const { sortReposByLastAccess } = useRepos();

const repoList = computed(() => sortReposByLastAccess(searchedRepos.value || []));

onMounted(async () => {
  await repoStore.loadRepos();
});
</script>
