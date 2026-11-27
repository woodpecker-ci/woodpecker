<template>
  <router-link
    v-if="repo"
    :to="{ name: 'repo', params: { repoId: repo.id } }"
    class="focus-visible:outline-wp-link-100 block min-w-0 rounded-xl focus-visible:outline-2 focus-visible:outline-offset-2"
  >
    <div
      class="border-wp-background-400 bg-wp-background-100 hover:border-wp-background-500 flex h-full flex-col overflow-hidden rounded-lg border shadow-xs transition-colors"
    >
      <div class="flex min-w-0 items-center gap-3 p-4">
        <span
          class="border-wp-background-400 bg-wp-background-200 grid size-9 shrink-0 place-items-center rounded-lg border"
        >
          <Icon name="repo" class="text-wp-text-100" />
        </span>
        <div class="min-w-0 flex-1">
          <div class="truncate text-sm font-semibold" :title="repo.name">{{ repo.name }}</div>
          <div class="text-wp-text-alt-100 truncate text-xs" :title="repo.owner">{{ repo.owner }}</div>
        </div>
        <Icon
          v-if="repo.visibility === RepoVisibility.Private"
          name="visibility-private"
          class="text-wp-text-alt-100 shrink-0"
          :title="$t('repo.visibility.private.private')"
        />
        <Icon
          v-else-if="repo.visibility === RepoVisibility.Internal"
          name="visibility-internal"
          class="text-wp-text-alt-100 shrink-0"
          :title="$t('repo.visibility.internal.internal')"
        />
        <Icon name="chevron-right" class="text-wp-text-alt-100 size-4! shrink-0" />
      </div>

      <div
        class="border-wp-background-400 bg-wp-background-200/65 text-wp-text-alt-100 flex min-h-15 flex-1 flex-col justify-center gap-2 border-t px-4 py-3 text-xs"
      >
        <template v-if="lastPipeline">
          <div class="flex min-w-0 flex-wrap items-center gap-x-3 gap-y-1">
            <span
              class="border-wp-background-400 bg-wp-background-100 text-wp-text-100 inline-flex max-w-full items-center gap-1.5 rounded-md border px-1.5 py-0.5 font-medium"
            >
              <PipelineStatusIcon :status="lastPipeline.status" show-label class="gap-1.5 [&_svg]:size-3.5!" />
            </span>
            <span class="ml-auto whitespace-nowrap">{{ since }}</span>
          </div>
          <RenderMarkdown
            class="overflow-hidden text-ellipsis whitespace-nowrap"
            :title="message"
            :content="shortMessage"
            inline
          />
        </template>
        <span v-else>{{ $t('repo.pipeline.no_pipelines') }}</span>
      </div>
    </div>
  </router-link>
</template>

<script lang="ts" setup>
import { computed } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import RenderMarkdown from '~/components/atomic/RenderMarkdown.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import type { Repo } from '~/lib/api/types';
import { RepoVisibility } from '~/lib/api/types';

const props = defineProps<{
  repo: Repo;
}>();

const lastPipeline = computed(() => props.repo.last_pipeline);
const { since, shortMessage, message } = usePipeline(lastPipeline);
</script>
