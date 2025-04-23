<template>
  <Panel>
    <ul class="w-full list-inside list-disc">
      <li v-for="file in pipeline.changed_files" :key="file">{{ file }}</li>
    </ul>
  </Panel>
</template>

<script lang="ts" setup>
import Panel from '~/components/layout/Panel.vue';
import { requiredInject } from '~/compositions/useInjectProvide';

const pipeline = requiredInject('pipeline');
</script>
