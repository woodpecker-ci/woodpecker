<template>
  <Panel v-if="pipeline">
    <div v-if="pipeline.changed_files === undefined || pipeline.changed_files.length < 1" class="w-full">
      <span class="text-wp-text-100">{{ $t('repo.pipeline.no_files') }}</span>
    </div>
    <div v-for="file in pipeline.changed_files" v-else :key="file" class="w-full">
      <div>- {{ file }}</div>
    </div>
  </Panel>
</template>

<script lang="ts" setup>
import { inject, Ref } from 'vue';

import Panel from '~/components/layout/Panel.vue';
import { Pipeline } from '~/lib/api/types';

const pipeline = inject<Ref<Pipeline>>('pipeline');
if (!pipeline) {
  throw new Error('Unexpected: "pipeline" should be provided at this place');
}
</script>
