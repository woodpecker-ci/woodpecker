<template>
  <div class="flex w-full mb-4 justify-center">
    <span class="text-color text-xl">{{ $t('repo.build.pipelines_for', { branch }) }}</span>
  </div>
  <BuildList :builds="builds" :repo="repo" />
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref, toRef } from 'vue';

import BuildList from '~/components/repo/build/BuildList.vue';
import { Build, Repo, RepoPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'RepoBranch',

  components: { BuildList },

  props: {
    branch: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const branch = toRef(props, 'branch');
    const repo = inject<Ref<Repo>>('repo');
    const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
    if (!repo || !repoPermissions) {
      throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
    }

    const allBuilds = inject<Ref<Build[]>>('builds');
    const builds = computed(() => allBuilds?.value.filter((b) => b.branch === branch.value));

    return { builds, repo };
  },
});
</script>
