<template>
  <Scaffold v-model:search="search" :go-back="goBack">
    <template #title>
      {{ $t('repo.add') }}
    </template>

    <div class="space-y-4">
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        class="items-center"
        :to="repo.active ? { name: 'repo', params: { repoId: repo.id } } : undefined"
      >
        <span class="text-color">{{ repo.full_name }}</span>
        <span v-if="repo.active" class="ml-auto text-color-alt">{{ $t('repo.enable.enabled') }}</span>
        <Button
          v-if="!repo.active"
          class="ml-auto"
          :text="$t('repo.enable.enable')"
          :is-loading="isActivatingRepo && repoToActivate?.id === repo.id"
          @click="activateRepo(repo)"
        />
      </ListItem>
    </div>
  </Scaffold>
</template>

<script lang="ts" setup>
import { onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';
import { Repo } from '~/lib/api/types';

const router = useRouter();
const apiClient = useApiClient();
const notifications = useNotifications();
const repos = ref<Repo[]>();
const repoToActivate = ref<Repo>();
const search = ref('');
const i18n = useI18n();

const { searchedRepos } = useRepoSearch(repos, search);

onMounted(async () => {
  repos.value = await apiClient.getRepoList({ all: true });
});

const { doSubmit: activateRepo, isLoading: isActivatingRepo } = useAsyncAction(async (repo: Repo) => {
  repoToActivate.value = repo;
  const _repo = await apiClient.activateRepo(repo.forge_remote_id);
  notifications.notify({ title: i18n.t('repo.enable.success'), type: 'success' });
  repoToActivate.value = undefined;
  await router.push({ name: 'repo', params: { repoId: _repo.id } });
});

const goBack = useRouteBackOrDefault({ name: 'repos' }, false);
</script>
