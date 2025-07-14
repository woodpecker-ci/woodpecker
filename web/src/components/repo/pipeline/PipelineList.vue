<template>
  <div class="space-y-4">
    <PipelineItem
      v-for="pipeline in pipelines"
      :key="pipeline.id"
      :to="{
        name: 'repo-pipeline',
        params: { pipelineId: pipeline.number },
      }"
      :pipeline="pipeline"
    />
    <div v-if="loading" class="flex justify-center">
      <Icon name="spinner" class="animate-spin" />
    </div>
    <Panel v-else-if="pipelines?.length === 0">
      <span class="text-wp-text-100">{{ $t('repo.pipeline.no_pipelines') }}</span>
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import Icon from '~/components/atomic/Icon.vue';
import Panel from '~/components/layout/Panel.vue';
import PipelineItem from '~/components/repo/pipeline/PipelineItem.vue';
import type { Pipeline } from '~/lib/api/types';

defineProps<{
  pipelines: Pipeline[] | undefined;
  loading?: boolean;
}>();
</script>
