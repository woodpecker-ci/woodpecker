<template>
  <FluidContainer class="flex flex-col">
    <div class="flex flex-row border-b pb-4 mb-4 items-center dark:border-dark-200">
      <h1 class="text-xl text-gray-500">Repositories</h1>
      <Button class="ml-auto" :to="{ name: 'repo-add' }" text="Add repository" />
    </div>

    <div class="space-y-4">
      <ListItem
        v-for="repo in repos"
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
import { defineComponent, onMounted } from 'vue';

import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'Repos',

  components: {
    Button,
    FluidContainer,
    ListItem,
  },

  setup() {
    const repoStore = RepoStore();
    const { repos } = repoStore;

    onMounted(async () => {
      await repoStore.loadRepos();
    });

    return { repos };
  },
});
</script>
