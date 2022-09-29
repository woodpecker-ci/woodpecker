<template>
  <div v-if="pipelines" class="space-y-4">
    <router-link
      v-for="pipeline in pipelines"
      :key="pipeline.id"
      :to="{
        name: 'repo-pipeline',
        params: { repoOwner: repo.owner, repoName: repo.name, pipelineId: pipeline.number },
      }"
      class="flex"
    >
      <PipelineItem :pipeline="pipeline" />
    </router-link>
    <Panel v-if="pipelines.length === 0">
      <span class="text-color">{{ $t('repo.pipeline.no_pipelines') }}</span>
    </Panel>
  </div>
</template>

<script lang="ts">
import { defineComponent, PropType } from 'vue';

import Panel from '~/components/layout/Panel.vue';
import PipelineItem from '~/components/repo/pipeline/PipelineItem.vue';
import { Pipeline, Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'PipelineList',

  components: { Panel, PipelineItem },

  props: {
    repo: {
      type: Object as PropType<Repo>,
      required: true,
    },

    pipelines: {
      type: Object as PropType<Pipeline[] | undefined>,
      required: true,
    },
  },
});
</script>
