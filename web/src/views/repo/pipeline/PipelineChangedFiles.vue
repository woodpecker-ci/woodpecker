<template>
  <Panel>
    <ul class="w-full list-inside list-disc">
      <li v-for="file in pipeline.changed_files" :key="file">{{ file }}</li>
    </ul>
  </Panel>
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import Panel from '~/components/layout/Panel.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';

const repo = requiredInject('repo');
const pipeline = requiredInject('pipeline');

const { t } = useI18n();
useWPTitle(
  computed(() => [
    t('repo.pipeline.files'),
    t('repo.pipeline.pipeline', { pipelineId: pipeline.value.number }),
    repo.value.full_name,
  ]),
);
</script>
