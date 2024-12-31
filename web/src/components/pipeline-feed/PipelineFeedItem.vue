<template>
  <div v-if="pipeline" class="text-wp-text-100 flex w-full">
    <PipelineStatusIcon :status="pipeline.status" class="flex items-center" />
    <div class="ml-4 flex min-w-0 flex-col">
      <router-link
        :to="{
          name: 'repo',
          params: { repoId: pipeline.repo_id },
        }"
        class="underline"
      >
        <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
        {{ repo?.owner }} / {{ repo?.name }}
      </router-link>
      <span class="overflow-hidden overflow-ellipsis whitespace-nowrap" :title="message">{{ shortMessage }}</span>
      <div class="mt-2 flex flex-col">
        <div class="flex items-center space-x-2" :title="created">
          <Icon name="since" />
          <span>{{ since }}</span>
        </div>
        <div class="flex items-center space-x-2">
          <Icon name="duration" />
          <span>{{ duration }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script lang="ts" setup>
import { computed, toRef } from 'vue';

import Icon from '~/components/atomic/Icon.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import usePipeline from '~/compositions/usePipeline';
import type { PipelineFeed } from '~/lib/api/types';
import { useRepoStore } from '~/store/repos';

const props = defineProps<{
  pipeline: PipelineFeed;
}>();

const repoStore = useRepoStore();

const pipeline = toRef(props, 'pipeline');
const repo = repoStore.getRepo(computed(() => pipeline.value.repo_id));

const { since, duration, shortMessage, message, created } = usePipeline(pipeline);
</script>
