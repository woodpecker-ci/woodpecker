<template>
  <router-link v-if="repo" :to="{ name: 'repo', params: { repoId: repo.id } }">
    <div
      class="flex flex-col border rounded-md bg-wp-background-100 overflow-hidden p-4 border-wp-background-400 dark:bg-wp-background-200 cursor-pointer hover:shadow-md hover:bg-wp-background-300 dark:hover:bg-wp-background-300"
    >
      <div class="grid grid-cols-[auto,1fr] gap-y-2 items-center">
        <div class="text-wp-text-100 text-lg">{{ `${repo.owner} / ${repo.name}` }}</div>
        <div class="ml-auto">
          <Badge
            v-if="repo.visibility === RepoVisibility.Public"
            :label="$t('repo.settings.general.visibility.public.public')"
          />
        </div>

        <div class="col-span-2 text-wp-text-100">
          <div v-if="pipeline" class="flex gap-x-1 items-center">
            <PipelineStatusIcon v-if="pipeline" :status="pipeline.status" />
            <Icon v-else name="new" class="text-wp-text-100" />
            <span class="whitespace-nowrap overflow-hidden overflow-ellipsis">{{ shortMessage }}</span>
          </div>

          <div v-else class="flex gap-x-2">
            <span>{{ $t('repo.pipeline.no_pipelines') }}</span>
          </div>
        </div>

        <div class="col-span-2 text-wp-text-100 text-sm min-h-5 mt-2">
          <div v-if="pipeline" class="flex gap-x-4 items-center">
            <div class="flex items-center gap-x-1">
              <span :title="pipelineEventTitle">
                <Icon v-if="pipeline.event === 'pull_request'" name="pull-request" size="22" />
                <Icon v-else-if="pipeline.event === 'pull_request_closed'" name="pull-request-closed" size="22" />
                <Icon v-else-if="pipeline.event === 'deployment'" name="deployment" size="22" />
                <Icon v-else-if="pipeline.event === 'tag' || pipeline.event === 'release'" name="tag" size="22" />
                <Icon v-else-if="pipeline.event === 'cron'" name="push" size="22" />
                <Icon v-else-if="pipeline.event === 'manual'" name="manual-pipeline" size="22" />
                <Icon v-else name="push" size="22" />
              </span>
              <span class="truncate">{{ prettyRef }}</span>
            </div>
            <div class="hidden sm:flex gap-x-1 items-center">
              <Icon name="commit" size="22" />
              <span class="truncate">{{ pipeline.commit.slice(0, 10) }}</span>
            </div>
            <div class="flex flex-shrink-0 gap-x-1 items-center ml-auto">
              <Icon name="since" size="22" />
              <span>{{ since }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </router-link>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';
import { useI18n } from 'vue-i18n';

import Badge from '~/components/atomic/Badge.vue';
import Icon from '~/components/atomic/Icon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import type { Repo } from '~/lib/api/types';
import { RepoVisibility } from '~/lib/api/types';

const props = defineProps<{
  repo: Repo;
}>();

const repo = props.repo;
const pipeline = toRef(repo.last_pipeline_item);
const { since, shortMessage, prettyRef } = usePipeline(pipeline);

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
