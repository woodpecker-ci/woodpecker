<template>
  <div class="space-y-4">
    <template v-if="pullRequests.length > 0">
      <ListItem
        v-for="pullRequest in pullRequests"
        :key="pullRequest.index"
        class="text-wp-text-100"
        :to="{ name: 'repo-pull-request', params: { pullRequest: pullRequest.index } }"
      >
        <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
        <span class="md:display-unset text-wp-text-alt-100 hidden">#{{ pullRequest.index }}</span>
        <!-- eslint-disable-next-line @intlify/vue-i18n/no-raw-text -->
        <span class="md:display-unset text-wp-text-alt-100 mx-2 hidden">-</span>
        <span class="text-wp-text-100 overflow-hidden text-ellipsis whitespace-nowrap underline md:no-underline">{{
          pullRequest.title
        }}</span>
      </ListItem>
    </template>
    <div v-else-if="loading" class="text-wp-text-100 flex justify-center">
      <Icon name="spinner" />
    </div>
    <Panel v-else class="flex justify-center">
      {{ $t('empty_list', { entity: $t('repo.pull_requests') }) }}
    </Panel>
  </div>
</template>

<script lang="ts" setup>
import { computed, watch } from 'vue';
import { useI18n } from 'vue-i18n';

import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Panel from '~/components/layout/Panel.vue';
import useApiClient from '~/compositions/useApiClient';
import { requiredInject } from '~/compositions/useInjectProvide';
import { usePagination } from '~/compositions/usePaginate';
import { useWPTitle } from '~/compositions/useWPTitle';
import type { PullRequest } from '~/lib/api/types';

const apiClient = useApiClient();

const repo = requiredInject('repo');
if (!repo.value.pr_enabled || !repo.value.allow_pr) {
  throw new Error('Unexpected: pull requests are disabled for repo');
}

async function loadPullRequests(page: number): Promise<PullRequest[]> {
  return apiClient.getRepoPullRequests(repo.value.id, { page });
}

const { resetPage, data: pullRequests, loading } = usePagination(loadPullRequests);

watch(repo, resetPage);

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.pull_requests'), repo.value.full_name]));
</script>
