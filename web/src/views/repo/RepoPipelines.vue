<template>
  <PipelineList
    :pipelines="pipelines"
    :loading="pipelineStore.loading"
    :has-more="pipelineStore.hasMore"
    @load-more="loadMore"
  />
</template>

<script lang="ts" setup>
import { computed, ref } from 'vue';
import { useI18n } from 'vue-i18n';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';
import { usePipelineStore } from '~/store/pipelines';

const repo = requiredInject('repo');
const pipelines = requiredInject('pipelines');
const pipelineStore = usePipelineStore();

const page = ref(1);

async function loadMore() {
  page.value += 1;
  await pipelineStore.loadRepoPipelines(repo.value.id, page.value);
}

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.activity'), repo.value.full_name]));
</script>
