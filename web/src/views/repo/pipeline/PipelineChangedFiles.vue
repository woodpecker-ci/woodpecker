<template>
  <Panel>
    <ul class="list-disc list-inside w-full">
      <li v-for="file in pipeline.changed_files" :key="file">{{ file }}</li>
    </ul>
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
