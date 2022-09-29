<template>
  <FluidContainer v-if="repo && repoPermissions && $route.meta.repoHeader">
    <div class="flex flex-wrap border-b items-center pb-4 mb-4 dark:border-gray-600 justify-center">
      <h1 class="text-xl text-color w-full md:w-auto text-center mb-4 md:mb-0">
        <router-link :to="{ name: 'repos-owner', params: { repoOwner } }" class="hover:underline">{{
          repoOwner
        }}</router-link>
        {{ ` / ${repo.name}` }}
      </h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="md:ml-auto">
        <img :src="badgeUrl" />
      </a>
      <a
        :href="repo.link_url"
        target="_blank"
        class="flex ml-4 p-1 rounded-full text-color hover:bg-gray-200 hover:text-gray-700 dark:hover:bg-gray-600"
      >
        <Icon v-if="forge === 'github'" name="github" />
        <Icon v-else-if="forge === 'gitea'" name="gitea" />
        <Icon v-else-if="forge === 'gitlab'" name="gitlab" />
        <Icon v-else-if="forge === 'bitbucket' || forge === 'stash'" name="bitbucket" />
        <Icon v-else name="repo" />
      </a>
      <IconButton v-if="repoPermissions.admin" class="ml-2" :to="{ name: 'repo-settings' }" icon="settings" />
    </div>

    <Tabs v-model="activeTab" disable-hash-mode class="mb-4">
      <Tab id="activity" :title="$t('repo.activity')" />
      <Tab id="branches" :title="$t('repo.branches')" />
    </Tabs>

    <router-view />
  </FluidContainer>
  <router-view v-else-if="repo && repoPermissions" />
</template>

<script lang="ts" setup>
import { computed, defineProps, onMounted, provide, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Tab from '~/components/tabs/Tab.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import useApiClient from '~/compositions/useApiClient';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import useNotifications from '~/compositions/useNotifications';
import { RepoPermissions } from '~/lib/api/types';
import BuildStore from '~/store/pipelines';
import RepoStore from '~/store/repos';

const props = defineProps({
  // used by toRef
  // eslint-disable-next-line vue/no-unused-properties
  repoOwner: {
    type: String,
    required: true,
  },

  // used by toRef
  // eslint-disable-next-line vue/no-unused-properties
  repoName: {
    type: String,
    required: true,
  },
});

const repoOwner = toRef(props, 'repoOwner');
const repoName = toRef(props, 'repoName');
const repoStore = RepoStore();
const buildStore = BuildStore();
const apiClient = useApiClient();
const notifications = useNotifications();
const { isAuthenticated } = useAuthentication();
const route = useRoute();
const router = useRouter();
const i18n = useI18n();

const { forge } = useConfig();
const repo = repoStore.getRepo(repoOwner, repoName);
const repoPermissions = ref<RepoPermissions>();
const builds = buildStore.getSortedPipelines(repoOwner, repoName);
provide('repo', repo);
provide('repo-permissions', repoPermissions);
provide('builds', builds);

async function loadRepo() {
  repoPermissions.value = await apiClient.getRepoPermissions(repoOwner.value, repoName.value);
  if (!repoPermissions.value.pull) {
    notifications.notify({ type: 'error', title: i18n.t('repo.not_allowed') });
    // no access and not authenticated, redirect to login
    if (!isAuthenticated) {
      await router.replace({ name: 'login', query: { url: route.fullPath } });
      return;
    }
    await router.replace({ name: 'home' });
    return;
  }

  const apiRepo = await repoStore.loadRepo(repoOwner.value, repoName.value);
  if (apiRepo.full_name !== `${repoOwner.value}/${repoName.value}`) {
    await router.replace({
      name: route.name ? route.name : 'repo',
      params: { repoOwner: apiRepo.owner, repoName: apiRepo.name },
    });
    return;
  }
  await buildStore.loadPipelines(repoOwner.value, repoName.value);
}

onMounted(() => {
  loadRepo();
});

watch([repoOwner, repoName], () => {
  loadRepo();
});

const badgeUrl = computed(() => `/api/badges/${repo.value.owner}/${repo.value.name}/status.svg`);

const activeTab = computed({
  get() {
    if (route.name === 'repo-branches' || route.name === 'repo-branch') {
      return 'branches';
    }
    return 'activity';
  },
  set(tab: string) {
    if (tab === 'branches') {
      router.push({ name: 'repo-branches' });
    } else {
      router.push({ name: 'repo' });
    }
  },
});
</script>
