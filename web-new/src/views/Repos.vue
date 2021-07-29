<template>
  <div class="max-w-5xl w-full h-full m-auto">
    <Button :to="{ name: 'repo-add' }" text="Add Repo" />

    <FluidContainer>
      <div
        @click="$router.push({ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } })"
        v-for="repo in repos"
      >
        {{ repo.full_name }}
      </div>
    </FluidContainer>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted } from 'vue';
import Button from '~/components/atomic/Button.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'Repos',

  components: {
    Button,
    FluidContainer,
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
