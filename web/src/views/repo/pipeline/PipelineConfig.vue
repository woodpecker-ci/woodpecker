<template>
  <div class="flex flex-col gap-y-4">
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
import { requiredInject } from '~/compositions/useInjectProvide';

const pipelineConfigs = requiredInject('pipeline-configs');

const pipelineConfigsDecoded = computed(
  () =>
    pipelineConfigs.value?.map((i) => ({
      ...i,
      data: decode(i.data),
    })) ?? [],
);
</script>
