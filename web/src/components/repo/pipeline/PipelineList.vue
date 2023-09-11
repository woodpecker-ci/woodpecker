<template>
  <div v-if="pipelines" class="space-y-4">
    <PipelineItem
      v-for="pipeline in pipelines"
      :key="pipeline.id"
      :to="{
        name: 'repo-pipeline',
        params: { pipelineId: pipeline.number },
      }"
      :pipeline="pipeline"
    />
    <Panel v-if="pipelines.length === 0">
      <span class="text-wp-text-100">{{ $t('repo.pipeline.no_pipelines') }}</span>
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import Panel from '~/components/layout/Panel.vue';
import PipelineItem from '~/components/repo/pipeline/PipelineItem.vue';
import { Pipeline } from '~/lib/api/types';

defineProps<{
  pipelines: Pipeline[] | undefined;
}>();
</script>
