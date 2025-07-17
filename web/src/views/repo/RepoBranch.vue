<template>
  <div class="mb-4 flex w-full justify-center">
    <span class="text-wp-text-100 text-xl">{{ $t('repo.pipeline.pipelines_for', { branch }) }}</span>
  </div>
  <PipelineList :pipelines="pipelines" :repo="repo" />
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';

const props = defineProps<{
  branch: string;
}>();

const branch = toRef(props, 'branch');
const repo = requiredInject('repo');

const allPipelines = requiredInject('pipelines');
const pipelines = computed(() =>
  allPipelines.value.filter(
    (b) =>
      b.branch === branch.value &&
      b.event !== 'pull_request' &&
      b.event !== 'pull_request_closed' &&
      b.event !== 'pull_request_metadata',
  ),
);

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.activity'), branch.value, repo.value.full_name]));
</script>
