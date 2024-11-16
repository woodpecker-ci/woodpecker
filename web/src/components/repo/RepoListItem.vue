<template>
  <router-link v-if="repo && pipeline" :to="{ name: 'repo', params: { repoId: repo.id } }">
    <div
      class="flex flex-col border rounded-md bg-wp-background-100 overflow-hidden p-4 border-wp-background-400 dark:bg-wp-background-200 cursor-pointer hover:shadow-md hover:bg-wp-background-300 dark:hover:bg-wp-background-300"
    >
      <div class="grid item grid-cols-[1.5rem,auto,1fr] gap-2">
        <PipelineStatusIcon :status="pipeline.status" class="flex items-center" />
        <span class="text-wp-text-100 text-lg">{{ `${repo.owner} / ${repo.name}` }}</span>
        <span class="ml-auto">
          <Badge
            v-if="repo.visibility === RepoVisibility.Public"
            :label="$t('repo.settings.general.visibility.public.public')"
          />
        </span>

        <div class="col-start-2 col-span-2 text-wp-text-100">
          <div v-if="pipeline" class="flex gap-2 items-center">
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
import { toRef } from 'vue';

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
</script>
