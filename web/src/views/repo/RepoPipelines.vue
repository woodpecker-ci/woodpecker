<template>
  <PipelineList :pipelines="pipelines" :repo="repo" />
</template>

<script lang="ts" setup>
import { computed } from 'vue';
import { useI18n } from 'vue-i18n';

import PipelineList from '~/components/repo/pipeline/PipelineList.vue';
import { requiredInject } from '~/compositions/useInjectProvide';
import { useWPTitle } from '~/compositions/useWPTitle';

const repo = requiredInject('repo');
const pipelines = requiredInject('pipelines');

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.activity'), repo.value.name]));
</script>
