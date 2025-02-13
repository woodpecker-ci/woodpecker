<template>
  <div class="flex flex-col gap-y-4">
    <Panel
      v-if="pipelineVariables && Object.keys(pipelineVariables).length > 0"
      collapsable
      collapsed-by-default
      :title="$t('repo.pipeline.variables')"
    >
      <div class="overflow-auto font-mono whitespace-pre">
        <div
          v-for="(value, key) in pipelineVariables"
          :key="key"
          class="border-wp-background-300 flex border-b py-2 last:border-b-0"
        >
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
      <SyntaxHighlight
        class="code-box overflow-auto font-mono whitespace-pre"
        language="yaml"
        :code="pipelineConfig.data"
      />
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import { decode } from 'js-base64';
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import SyntaxHighlight from '~/components/atomic/SyntaxHighlight';
import Panel from '~/components/layout/Panel.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';

const repo = requiredInject('repo');
const pipeline = requiredInject('pipeline');
const pipelineConfigs = requiredInject('pipeline-configs');
const pipelineVariables = requiredInject('pipeline-variables');

const pipelineConfigsDecoded = computed(
  () =>
    pipelineConfigs.value?.map((i) => ({
      ...i,
      data: decode(i.data),
    })) ?? [],
);

const { t } = useI18n();
useWPTitle(
  computed(() => [
    t('repo.pipeline.config'),
    t('repo.pipeline.pipeline', { pipelineId: pipeline.value.number }),
    repo.value.full_name,
  ]),
);
</script>
