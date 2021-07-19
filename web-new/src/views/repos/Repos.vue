<template>
  <div class="max-w-5xl w-full h-full m-auto">
    <Button :to="{ name: 'repo-add' }" text="Add Repo" />
    <div
      @click="$router.push({ name: 'repo', params: { repoId: repo.name, repoOwner: repo.owner } })"
      v-for="repo in repos"
    >
      {{ repo.full_name }}
    </div>
  </div>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';
import Button from '~/components/atomic/Button.vue';

export default defineComponent({
  name: 'Repos',

  components: {
    Button,
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
