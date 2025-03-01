<template>
  <div class="flex flex-col gap-y-6">
    <Panel
      v-if="pipelineVariables && Object.keys(pipelineVariables).length > 0"
      collapsable
      collapsed-by-default
      :title="$t('repo.pipeline.variables')"
    >
      <div class="overflow-auto whitespace-pre font-mono">
        <div v-for="(value, key) in pipelineVariables" :key="key" class="border-b border-wp-background-300 py-2 last:border-b-0 flex">
          <span class="text-wp-text-100 min-w-[100px]">{{ key }}:</span>
          <span class="text-wp-text-100 flex-1">{{ value }}</span>
        </div>
      </div>
    </Panel>

    <Panel
      v-for="pipelineConfig in pipelineConfigsDecoded"
      :key="pipelineConfig.hash"
      :collapsable="pipelineConfigsDecoded && pipelineConfigsDecoded.length > 1"
      collapsed-by-default
      :title="pipelineConfigsDecoded && pipelineConfigsDecoded.length > 1 ? pipelineConfig.name : ''"
    >
      <SyntaxHighlight class="overflow-auto font-mono whitespace-pre" language="yaml" :code="pipelineConfig.data" />
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import { decode } from 'js-base64';
import { computed } from 'vue';

import SyntaxHighlight from '~/components/atomic/SyntaxHighlight';
import Panel from '~/components/layout/Panel.vue';
import { inject } from '~/compositions/useInjectProvide';

const pipelineConfigs = inject('pipeline-configs');
if (!pipelineConfigs) {
  throw new Error('Unexpected: "pipelineConfigs" should be provided at this place');
}
const pipelineVariables = inject('pipeline-variables');

const pipelineConfigsDecoded = computed(
  () =>
    pipelineConfigs.value?.map((i) => ({
      ...i,
      data: decode(i.data),
    })) ?? [],
);
</script>
