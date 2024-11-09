<template>
  <Scaffold v-model:search="search" :go-back="goBack">
    <template #title>
      {{ $t('repo.add') }}
    </template>

    <div class="space-y-4">
      <template v-if="repos !== undefined && repos.length > 0">
        <ListItem
          v-for="repo in searchedRepos"
          :key="repo.id"
          class="items-center"
          :to="repo.active ? { name: 'repo', params: { repoId: repo.id } } : undefined"
        >
          <span class="text-wp-text-100">{{ repo.full_name }}</span>
          <span v-if="repo.active" class="ml-auto text-wp-text-alt-100">{{ $t('repo.enable.enabled') }}</span>
          <div v-else class="ml-auto flex items-center">
            <Badge v-if="repo.id" class="<md:hidden mr-2" :label="$t('repo.enable.disabled')" />
            <Button
              :text="$t('repo.enable.enable')"
              :is-loading="isActivatingRepo && repoToActivate?.forge_remote_id === repo.forge_remote_id"
              @click="activateRepo(repo)"
            />
          </div>
        </ListItem>
      </template>
      <div v-else-if="loading" class="flex justify-center text-wp-text-100">
        <Icon name="spinner" />
      </div>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Badge from '~/components/atomic/Badge.vue';
import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { useRouteBack } from '~/compositions/useRouteBack';
import type { Repo } from '~/lib/api/types';

const router = useRouter();
const apiClient = useApiClient();
const notifications = useNotifications();
const repos = ref<Repo[]>();
const repoToActivate = ref<Repo>();
const search = ref('');
const i18n = useI18n();
const loading = ref(false);

const { searchedRepos } = useRepoSearch(repos, search);

onMounted(async () => {
  loading.value = true;
  repos.value = await apiClient.getRepoList({ all: true });
  loading.value = false;
});

const { doSubmit: activateRepo, isLoading: isActivatingRepo } = useAsyncAction(async (repo: Repo) => {
  repoToActivate.value = repo;
  const _repo = await apiClient.activateRepo(repo.forge_remote_id);
  notifications.notify({ title: i18n.t('repo.enable.success'), type: 'success' });
  repoToActivate.value = undefined;
  await router.push({ name: 'repo', params: { repoId: _repo.id } });
});

const goBack = useRouteBack({ name: 'repos' });
</script>
