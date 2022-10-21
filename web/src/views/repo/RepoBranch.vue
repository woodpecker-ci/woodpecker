<template>
  <div class="flex w-full mb-4 justify-center">
    <span class="text-color text-xl">{{ $t('repo.pipeline.pipelines_for', { branch }) }}</span>
  </div>
  <PipelineList :pipelines="pipelines" :repo="repo" />
</template>

<script lang="ts">
import { computed, defineComponent, inject, Ref, toRef } from 'vue';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { Pipeline, Repo, RepoPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'RepoBranch',

  components: { PipelineList },

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

    const allPipelines = inject<Ref<Pipeline[]>>('pipelines');
    const pipelines = computed(() => allPipelines?.value.filter((b) => b.branch === branch.value));

    return { pipelines, repo };
  },
});
</script>
