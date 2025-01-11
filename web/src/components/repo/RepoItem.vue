<template>
  <router-link
    v-if="repo"
    :to="{ name: 'repo', params: { repoId: repo.id } }"
    class="flex cursor-pointer flex-col overflow-hidden rounded-md border border-wp-background-400 bg-wp-background-100 p-4 hover:bg-wp-background-300 hover:shadow-md dark:bg-wp-background-200 dark:hover:bg-wp-background-300"
  >
    <div class="grid grid-cols-[auto,1fr] items-center gap-y-4">
      <div class="text-lg text-wp-text-100">{{ `${repo.owner} / ${repo.name}` }}</div>
      <div class="ml-auto text-wp-text-100">
        <div
          v-if="repo.visibility === RepoVisibility.Private"
          :title="`${$t('repo.visibility.visibility')}: ${$t(`repo.visibility.private.private`)}`"
        >
          <Icon name="visibility-private" />
        </div>
        <div
          v-else-if="repo.visibility === RepoVisibility.Internal"
          :title="`${$t('repo.visibility.visibility')}: ${$t(`repo.visibility.internal.internal`)}`"
        >
          <Icon name="visibility-internal" />
        </div>
      </div>

      <div class="col-span-2 flex w-full gap-x-4 text-wp-text-100">
        <template v-if="lastPipeline">
          <div class="flex min-w-0 flex-1 items-center gap-x-1">
            <PipelineStatusIcon v-if="lastPipeline" :status="lastPipeline.status" />
            <span class="overflow-hidden overflow-ellipsis whitespace-nowrap pl-1">{{ shortMessage }}</span>
          </div>

          <div class="ml-auto flex flex-shrink-0 items-center gap-x-1">
            <Icon name="since" />
            <span>{{ since }}</span>
          </div>
        </template>

        <div v-else class="flex gap-x-2">
          <span>{{ $t('repo.pipeline.no_pipelines') }}</span>
        </div>
      </div>
    </div>
  </router-link>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import type { Repo } from '~/lib/api/types';
import { RepoVisibility } from '~/lib/api/types';

const props = defineProps<{
  repo: Repo;
}>();

const lastPipeline = computed(() => props.repo.last_pipeline_item);
const { since, shortMessage } = usePipeline(lastPipeline);
</script>
