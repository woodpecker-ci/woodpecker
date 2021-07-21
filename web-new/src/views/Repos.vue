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
import { defineComponent, onMounted, ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';

export default defineComponent({
  name: 'Repos',

  components: {
    Button,
    FluidContainer,
  },

  setup() {
    const apiClient = useApiClient();
    const repos = ref<Repo[] | undefined>();

    onMounted(async () => {
      repos.value = await apiClient.getRepoList();
    });

    return { repos };
  },
});
</script>
