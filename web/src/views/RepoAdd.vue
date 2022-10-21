<template>
  <FluidContainer class="flex flex-col">
    <div class="flex flex-row border-b mb-4 pb-4 items-center dark:border-dark-200">
      <IconButton :to="{ name: 'repos' }" :title="$t('back')" icon="back" />
      <h1 class="text-xl ml-2 text-color">{{ $t('repo.add') }}</h1>
      <TextField v-model="search" class="w-auto ml-auto" :placeholder="$t('search')" />
      <Button
        class="ml-auto"
        start-icon="sync"
        :text="$t('repo.enable.reload')"
        :is-loading="isReloadingRepos"
        @click="reloadRepos"
      />
    </div>

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
  </FluidContainer>
</template>

<script lang="ts">
import { defineComponent, onMounted, ref } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ListItem from '~/components/atomic/ListItem.vue';
import TextField from '~/components/form/TextField.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { useRepoSearch } from '~/compositions/useRepoSearch';
import { Repo } from '~/lib/api/types';

export default defineComponent({
  name: 'RepoAdd',

  components: {
    Button,
    FluidContainer,
    ListItem,
    IconButton,
    TextField,
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

    return {
      isReloadingRepos,
      isActivatingRepo,
      repoToActivate,
      reloadRepos,
      activateRepo,
      searchedRepos,
      search,
    };
  },
});
</script>
