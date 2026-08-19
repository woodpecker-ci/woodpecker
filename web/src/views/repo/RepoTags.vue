<template>
  <div class="space-y-4">
    <ListItem
      v-for="tag in tags"
      :key="tag.name"
      class="text-wp-text-100"
      :to="{ name: 'repo-tag', params: { tag: tag.name } }"
    >
      {{ tag.name }}
    </ListItem>
    <div v-if="loading" class="text-wp-text-100 flex justify-center">
      <Icon name="spinner" />
    </div>
    <Panel v-else-if="tags.length === 0" class="flex justify-center">
      {{ $t('empty_list', { entity: $t('repo.tags') }) }}
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
import type { RepoTag } from '~/lib/api/types';

const apiClient = useApiClient();

const repo = requiredInject('repo');

async function loadTags(page: number): Promise<RepoTag[]> {
  return apiClient.getRepoTags(repo.value.id, { page });
}

const { resetPage, data: tags, loading } = usePagination(loadTags);

watch(repo, resetPage);

const { t } = useI18n();
useWPTitle(computed(() => [t('repo.tags'), repo.value.full_name]));
</script>
