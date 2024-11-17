<template>
  <router-link v-if="repo" :to="{ name: 'repo', params: { repoId: repo.id } }">
    <div
      class="flex flex-col border rounded-md bg-wp-background-100 overflow-hidden p-4 border-wp-background-400 dark:bg-wp-background-200 cursor-pointer hover:shadow-md hover:bg-wp-background-300 dark:hover:bg-wp-background-300"
    >
      <div class="grid grid-cols-[1.5rem,auto,1fr] gap-2 items-center">
        <PipelineStatusIcon v-if="pipeline" :status="pipeline.status" />
        <Icon v-else name="new" class="text-wp-text-100" />
        <span class="text-wp-text-100 text-lg">{{ `${repo.owner} / ${repo.name}` }}</span>
        <span class="ml-auto">
          <Badge
            v-if="repo.visibility === RepoVisibility.Public"
            :label="$t('repo.settings.general.visibility.public.public')"
          />
        </span>

        <div class="col-start-2 col-span-2 text-wp-text-100">
          <div v-if="pipeline" class="flex gap-1 items-center">
            <span :title="pipelineEventTitle">
              <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" />
              <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" />
              <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" />
              <Icon v-else-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" />
              <Icon v-else-if="pipeline.event === 'cron'" name="push" />
              <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" />
              <Icon v-else name="push" />
            </span>
            <span class="whitespace-nowrap overflow-hidden overflow-ellipsis">{{ shortMessage }}</span>
            <span class="ml-auto hidden sm:inline-block flex-shrink-0">{{ since }}</span>
          </div>

          <div v-else class="flex gap-2">
            <span>{{ $t('repo.pipeline.no_pipelines') }}</span>
          </div>
        </div>
      </div>
    </div>
  </router-link>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import Badge from '~/components/atomic/Badge.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import type { Repo } from '~/lib/api/types';
import { RepoVisibility } from '~/lib/api/types';

const props = defineProps<{
  repo: Repo;
}>();

const repo = props.repo;
const pipeline = toRef(repo.last_pipeline);
const { since, shortMessage } = usePipeline(pipeline);

const { t } = useI18n();

const pipelineEventTitle = computed(() => {
  switch (pipeline.value.event) {
    case 'pull_request':
      return t('repo.pipeline.event.pr');
    case 'pull_request_closed':
      return t('repo.pipeline.event.pr_closed');
    case 'deployment':
      return t('repo.pipeline.event.deploy');
    case 'tag':
      return t('repo.pipeline.event.tag');
    case 'release':
      return t('repo.pipeline.event.release');
    case 'cron':
      return t('repo.pipeline.event.cron');
    case 'manual':
      return t('repo.pipeline.event.manual');
    default:
      return t('repo.pipeline.event.push');
  }
});
</script>
