<template>
  <div v-if="branches" class="space-y-4">
    <ListItem
      v-for="branch in branches"
      :key="branch"
      class="text-color"
      :to="{ name: 'repo-branch', params: { branch } }"
    >
      {{ branch }}
    </ListItem>
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

      branches.value = await apiClient.getRepoBranches(repo.value.id);
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
