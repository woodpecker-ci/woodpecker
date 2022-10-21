<template>
  <Scaffold :go-back="goBack" three-column-header>
    <template #headerTitle>
      {{ $t('repo.add') }}
    </template>

    <template #headerCenterBox>
      <TextField v-model="search" class="w-auto !bg-gray-100 !dark:bg-dark-gray-600" :placeholder="$t('search')" />
    </template>

    <template #headerActions>
      <Button start-icon="sync" :text="$t('repo.enable.reload')" :is-loading="isReloadingRepos" @click="reloadRepos" />
    </template>

    <div class="space-y-4">
      <ListItem
        v-for="repo in searchedRepos"
        :key="repo.id"
        class="items-center"
        :clickable="repo.active"
        @click="repo.active && $router.push({ name: 'repo', params: { repoOwner: repo.owner, repoName: repo.name } })"
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
import TextField from '~/components/form/TextField.vue';
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
    TextField,
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
      notifications.notify({ title: i18n.t('repo.enabled.success'), type: 'success' });
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
