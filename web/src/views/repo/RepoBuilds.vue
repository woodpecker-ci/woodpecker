<template>
  <BuildList :builds="builds" :repo="repo" />
</template>

<script lang="ts">
import { defineComponent, inject, Ref } from 'vue';

import BuildList from '~/components/repo/build/BuildList.vue';
import { Build, Repo, RepoPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'RepoBuilds',

  components: { BuildList },

  setup() {
    const repo = inject<Ref<Repo>>('repo');
    const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
    if (!repo || !repoPermissions) {
      throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
    }

    const builds = inject<Ref<Build[]>>('builds');

    return { builds, repo };
  },
});
</script>
