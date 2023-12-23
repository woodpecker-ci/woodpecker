<template>
  <Scaffold
    v-if="repo && repoPermissions && $route.meta.repoHeader"
    v-model:activeTab="activeTab"
    enable-tabs
    disable-tab-url-hash-mode
  >
    <template #title>
      <span class="flex">
        <router-link :to="{ name: 'org', params: { orgId: repo.org_id } }" class="hover:underline">{{
          repo.owner
        }}</router-link>
        {{ `&nbsp;/&nbsp;${repo.name}` }}
      </span>
    </template>
    <template #titleActions>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank">
        <img :src="badgeUrl" />
      </a>
      <IconButton :href="repo.forge_url" :title="$t('repo.open_in_forge')" :icon="forge ?? 'repo'" />
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
        :text="$t('repo.manual_pipeline.trigger')"
        @click="showManualPipelinePopup = true"
      />
      <ManualPipelinePopup :open="showManualPipelinePopup" @close="showManualPipelinePopup = false" />
    </template>

    <Tab id="activity" :title="$t('repo.activity')" />
    <Tab id="branches" :title="$t('repo.branches')" />
    <Tab id="pull_requests" :title="$t('repo.pull_requests')" />

    <router-view />
  </Scaffold>
  <router-view v-else-if="repo && repoPermissions" />
</template>

<script lang="ts" setup>
import { computed, onMounted, provide, ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import IconButton from '~/components/atomic/IconButton.vue';
import ManualPipelinePopup from '~/components/layout/popups/ManualPipelinePopup.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import useApiClient from '~/compositions/useApiClient';
import useAuthentication from '~/compositions/useAuthentication';
import useConfig from '~/compositions/useConfig';
import useNotifications from '~/compositions/useNotifications';
import { RepoPermissions } from '~/lib/api/types';
import { usePipelineStore } from '~/store/pipelines';
import { useRepoStore } from '~/store/repos';

const props = defineProps<{
  repoId: string;
}>();

const _repoId = toRef(props, 'repoId');
const repositoryId = computed(() => parseInt(_repoId.value, 10));
const repoStore = useRepoStore();
const pipelineStore = usePipelineStore();
const apiClient = useApiClient();
const notifications = useNotifications();
const { isAuthenticated } = useAuthentication();
const route = useRoute();
const router = useRouter();
const i18n = useI18n();
const config = useConfig();

const { forge } = useConfig();
const repo = repoStore.getRepo(repositoryId);
const repoPermissions = ref<RepoPermissions>();
const pipelines = pipelineStore.getRepoPipelines(repositoryId);
provide('repo', repo);
provide('repo-permissions', repoPermissions);
provide('pipelines', pipelines);

const showManualPipelinePopup = ref(false);

async function loadRepo() {
  repoPermissions.value = await apiClient.getRepoPermissions(repositoryId.value);
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

  await repoStore.loadRepo(repositoryId.value);
  await pipelineStore.loadRepoPipelines(repositoryId.value);
}

onMounted(() => {
  loadRepo();
});

watch([repositoryId], () => {
  loadRepo();
});

const badgeUrl = computed(() => repo.value && `${config.rootPath}/api/badges/${repo.value.id}/status.svg`);

const activeTab = computed({
  get() {
    if (route.name === 'repo-branches' || route.name === 'repo-branch') {
      return 'branches';
    }
    if (route.name === 'repo-pull-requests' || route.name === 'repo-pull-request') {
      return 'pull_requests';
    }
    return 'activity';
  },
  set(tab: string) {
    if (tab === 'branches') {
      router.push({ name: 'repo-branches' });
    } else if (tab === 'pull_requests') {
      router.push({ name: 'repo-pull-requests' });
    } else {
      router.push({ name: 'repo' });
    }
  },
});
</script>
