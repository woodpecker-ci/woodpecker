<template>
  <FluidContainer class="flex flex-col">
    <div class="flex flex-row flex-wrap md:grid md:grid-cols-3 border-b pb-4 mb-4 dark:border-dark-200">
      <h1 class="text-xl text-gray-500">{{ repoOwner }}</h1>
      <TextField v-model="search" class="w-auto md:ml-auto md:mr-auto" :placeholder="$t('search')" />
    </div>

    <div class="space-y-4">
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        clickable
        @click="$router.push({ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } })"
      >
        <span class="text-gray-500">{{ `${repo.name}` }}</span>
      </ListItem>
    </div>
    <div v-if="(searchedRepos || []).length <= 0" class="text-center">
      <span class="text-gray-500 m-auto">{{ $t('repo.user_none') }}</span>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, ref } from 'vue';

import ListItem from '~/components/atomic/ListItem.vue';
import TextField from '~/components/form/TextField.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'ReposOwner',

  components: {
    FluidContainer,
    ListItem,
    TextField,
  },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const repoStore = RepoStore();
    // TODO: filter server side
    const repos = computed(() => Object.values(repoStore.repos).filter((v) => v.owner === props.repoOwner));
    const search = ref('');

    const { searchedRepos } = useRepoSearch(repos, search);

    onMounted(async () => {
      await repoStore.loadRepos();
    });

    return { searchedRepos, search };
  },
});
</script>
