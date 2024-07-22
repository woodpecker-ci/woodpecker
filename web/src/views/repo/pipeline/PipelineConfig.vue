<template>
  <div class="flex flex-col gap-y-6">
    <Panel
      v-for="pipelineConfig in pipelineConfigsDecoded || []"
      :key="pipelineConfig.hash"
      :collapsable="pipelineConfigsDecoded && pipelineConfigsDecoded.length > 1"
      collapsed-by-default
      :title="pipelineConfigsDecoded && pipelineConfigsDecoded.length > 1 ? pipelineConfig.name : ''"
    >
      <SyntaxHighlight class="font-mono whitespace-pre overflow-auto" language="yaml" :code="pipelineConfig.data" />
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import { decode } from 'js-base64';
import { computed, inject, type Ref } from 'vue';

import SyntaxHighlight from '~/components/atomic/SyntaxHighlight';
import Panel from '~/components/layout/Panel.vue';
import type { PipelineConfig } from '~/lib/api/types';

const pipelineConfigs = inject<Ref<PipelineConfig[]>>('pipeline-configs');
if (!pipelineConfigs) {
  throw new Error('Unexpected: "pipelineConfigs" should be provided at this place');
}

const pipelineConfigsDecoded = computed(() =>
  pipelineConfigs.value.map((i) => ({
    ...i,
    data: decode(i.data),
  })),
);
</script>
