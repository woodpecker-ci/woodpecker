<template>
  <FluidContainer class="flex flex-col">
    <div class="flex flex-row flex-wrap md:grid md:grid-cols-3 border-b pb-4 mb-4 dark:border-dark-200">
      <h1 class="text-xl text-gray-500">Repositories</h1>
      <TextField v-model="search" class="w-auto md:ml-auto md:mr-auto" :placeholder="$t('search')" />
      <Button class="md:ml-auto" :to="{ name: 'repo-add' }" start-icon="plus" :text="$t('repo.add')" />
    </div>

    <div class="space-y-4">
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        clickable
        @click="$router.push({ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } })"
      >
        <span class="text-gray-500">{{ `${repo.owner} / ${repo.name}` }}</span>
      </ListItem>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import TextField from '~/components/form/TextField.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'Repos',

  components: {
    Button,
    FluidContainer,
    ListItem,
    TextField,
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
