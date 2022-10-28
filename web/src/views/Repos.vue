<template>
  <Scaffold v-model:search="search">
    <template #title>
      {{ $t('repositories') }}
    </template>

    <template #titleActions>
      <Button :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </template>

    <div class="space-y-4">
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        :to="{ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } }"
      >
        <span class="text-color">{{ `${repo.owner} / ${repo.name}` }}</span>
      </ListItem>
    </div>
  </Scaffold>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'Repos',

  components: {
    Button,
    ListItem,
    Scaffold,
  },

  setup() {
    const repoStore = RepoStore();
    const repos = computed(() => Object.values(repoStore.repos));
    const search = ref('');

    const { searchedRepos } = useRepoSearch(repos, search);

    onMounted(async () => {
      await repoStore.loadRepos();
    });

    return { searchedRepos, search };
  },
});
</script>
