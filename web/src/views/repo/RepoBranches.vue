<template>
  <div v-if="branches" class="space-y-4">
    <router-link
      v-for="branch in branches"
      :key="branch"
      :to="{ name: 'repo-branch', params: { branch } }"
      class="flex"
    >
      <ListItem clickable class="text-gray-600 dark:text-gray-500">
        {{ branch }}
      </ListItem>
    </router-link>
  </div>
</template>

<script lang="ts">
import { defineComponent, inject, onMounted, Ref, ref, watch } from 'vue';

import ListItem from '~/components/atomic/ListItem.vue';
import useApiClient from '~/compositions/useApiClient';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'RepoBranches',

  components: {
    ListItem,
  },

  setup() {
    const apiClient = useApiClient();

    const branches = ref<string[]>();
    const repo = inject<Ref<Repo>>('repo');
    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    async function loadBranches() {
      if (!repo) {
        throw new Error('Unexpected: "repo" should be provided at this place');
      }

      branches.value = await apiClient.getRepoBranches(repo.value.owner, repo.value.name);
    }

    onMounted(() => {
      loadBranches();
    });

    watch(repo, () => {
      loadBranches();
    });

    return { branches };
  },
});
</script>
