<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('repositories') }}
    </template>

    <template #titleActions>
      <Button :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </template>

    <div class="space-y-4">
      <ListItem v-for="repo in searchedRepos" :key="repo.id" :to="{ name: 'repo', params: { repoId: repo.id } }">
        <span class="text-wp-text-100">{{ `${repo.owner} / ${repo.name}` }}</span>
      </ListItem>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { computed, onMounted, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { useRepoStore } from '~/store/repos';

const repoStore = useRepoStore();
const repos = computed(() => Object.values(repoStore.ownedRepos));
const search = ref('');

const { searchedRepos } = useRepoSearch(repos, search);

onMounted(async () => {
  await repoStore.loadRepos();
});
</script>
