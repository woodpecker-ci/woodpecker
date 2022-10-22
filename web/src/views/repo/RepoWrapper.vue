<template>
  <Scaffold
    v-if="repo && repoPermissions && $route.meta.repoHeader"
    v-model:activeTab="activeTab"
    enable-tabs
    disable-hash-mode
  >
    <template #title>
      <span class="flex">
        <router-link :to="{ name: 'repos-owner', params: { repoOwner } }" class="hover:underline">{{
          repoOwner
        }}</router-link>
        {{ `&nbsp;/&nbsp;${repo.name}` }}
      </span>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-2">
        <img :src="badgeUrl" />
      </a>
    </template>
    <template #titleActions>
      <a
        :href="repo.link_url"
        target="_blank"
        class="flex p-1 rounded-full text-color hover:bg-gray-200 hover:text-gray-700 dark:hover:bg-gray-600"
      >
        <Icon v-if="forge === 'github'" name="github" />
        <Icon v-else-if="forge === 'gitea'" name="gitea" />
        <Icon v-else-if="forge === 'gitlab'" name="gitlab" />
        <Icon v-else-if="forge === 'bitbucket' || forge === 'stash'" name="bitbucket" />
        <Icon v-else name="repo" />
      </a>
      <IconButton
        v-if="repoPermissions.admin"
        :to="{ name: 'repo-settings' }"
        :title="$t('repo.settings.settings')"
        icon="settings"
      />
    </template>

    <template #tabActions>
      <Button
        v-if="repoPermissions.push"
        type="submit"
        :text="$t('repo.manual_pipeline.trigger')"
        @click="showManualPipelinePopup = true"
      />
      <ManualPipelinePopup :open="showManualPipelinePopup" @close="showManualPipelinePopup = false" />
    </template>

    <Tab id="activity" :title="$t('repo.activity')" />
    <Tab id="branches" :title="$t('repo.branches')" />

    <router-view />
  </Scaffold>
  <router-view v-else-if="repo && repoPermissions" />
</template>

<script lang="ts" setup>
import { computed, onMounted, provide, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import ManualPipelinePopup from '~/components/layout/popups/ManualPipelinePopup.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useApiClient from '~/compositions/useApiClient';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import useNotifications from '~/compositions/useNotifications';
import { RepoPermissions } from '~/lib/api/types';
import PipelineStore from '~/store/pipelines';
import RepoStore from '~/store/repos';

const props = defineProps({
  repoOwner: {
    type: String,
    required: true,
  },

  repoName: {
    type: String,
    required: true,
  },
});

const repoOwner = toRef(props, 'repoOwner');
const repoName = toRef(props, 'repoName');
const repoStore = RepoStore();
const pipelineStore = PipelineStore();
const apiClient = useApiClient();
const notifications = useNotifications();
const { isAuthenticated } = useAuthentication();
const route = useRoute();
const router = useRouter();
const i18n = useI18n();

const { forge } = useConfig();
const repo = repoStore.getRepo(repoOwner, repoName);
const repoPermissions = ref<RepoPermissions>();
const pipelines = pipelineStore.getSortedPipelines(repoOwner, repoName);
provide('repo', repo);
provide('repo-permissions', repoPermissions);
provide('pipelines', pipelines);

const showManualPipelinePopup = ref(false);

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
  await pipelineStore.loadPipelines(repoOwner.value, repoName.value);
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
