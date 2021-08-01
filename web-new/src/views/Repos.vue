<template>
  <FluidContainer class="flex flex-col">
    <div class="flex flex-row border-b pb-4 mb-4 items-center">
      <h1 class="text-xl">Repositories</h1>
      <Button class="ml-auto" :to="{ name: 'repo-add' }" text="Add Repo" />
    </div>

    <div class="space-y-4">
      <ListItem
        v-for="repo in repos"
        :key="repo.id"
        clickable
        @click="$router.push({ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } })"
      >
        {{ repo.full_name }}
      </ListItem>
    </div>
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, onMounted } from 'vue';
import Button from '~/components/atomic/Button.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import RepoStore from '~/store/repos';
import ListItem from '~/components/atomic/ListItem.vue';

export default defineComponent({
  name: 'Repos',

  components: {
    Button,
    FluidContainer,
    ListItem,
  },

  setup() {
    const repoStore = RepoStore();
    const repos = repoStore.repos;

    onMounted(async () => {
      await repoStore.loadRepos();
    });

    return { repos };
  },
});
</script>
