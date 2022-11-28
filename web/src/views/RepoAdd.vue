<template>
  <Scaffold v-model:search="search" :go-back="goBack">
    <template #title>
      {{ $t('repo.add') }}
    </template>

    <template #titleActions>
      <Button start-icon="sync" :text="$t('repo.enable.reload')" :is-loading="isReloadingRepos" @click="reloadRepos" />
    </template>

    <div class="space-y-4">
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        class="items-center"
        :to="repo.active ? { name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } } : undefined"
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

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
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

export default defineComponent({
  name: 'RepoAdd',

  components: {
    Button,
    ListItem,
    Scaffold,
  },

  setup() {
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

    const { doSubmit: reloadRepos, isLoading: isReloadingRepos } = useAsyncAction(async () => {
      repos.value = undefined;
      repos.value = await apiClient.getRepoList({ all: true, flush: true });
      notifications.notify({ title: i18n.t('repo.enable.list_reloaded'), type: 'success' });
    });

    const { doSubmit: activateRepo, isLoading: isActivatingRepo } = useAsyncAction(async (repo: Repo) => {
      repoToActivate.value = repo;
      await apiClient.activateRepo(repo.owner, repo.name);
      notifications.notify({ title: i18n.t('repo.enable.success'), type: 'success' });
      repoToActivate.value = undefined;
      await router.push({ name: 'repo', params: { repoName: repo.name, repoOwner: repo.owner } });
    });

    const goBack = useRouteBackOrDefault({ name: 'repos' });

    return {
      isReloadingRepos,
      isActivatingRepo,
      repoToActivate,
      goBack,
      reloadRepos,
      activateRepo,
      searchedRepos,
      search,
    };
  },
});
</script>
