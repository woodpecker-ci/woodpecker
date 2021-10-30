<template>
  <FluidContainer v-if="repo && repoPermissions && $route.meta.repoHeader">
    <div class="flex border-b items-center pb-4 mb-4 dark:border-gray-600">
      <h1 class="text-xl text-gray-500">{{ `${repo.owner} / ${repo.name}` }}</h1>
      <a v-if="badgeUrl" :href="badgeUrl" target="_blank" class="ml-auto">
        <img :src="badgeUrl" />
      </a>
      <a
        :href="repo.link_url"
        target="_blank"
        class="flex ml-4 p-1 rounded-full text-gray-500 hover:bg-gray-200 hover:text-gray-700 dark:hover:bg-gray-600"
      >
        <Icon v-if="repo.link_url.startsWith('https://github.com/')" name="github" />
        <Icon v-else name="repo" />
      </a>
      <IconButton v-if="repoPermissions.admin" class="ml-2" :to="{ name: 'repo-settings' }" icon="settings" />
    </div>

    <Tabs v-model="activeTab" disable-hash-mode>
      <Tab title="Activity" />
      <Tab title="Branches" />
    </Tabs>

    <router-view />
  </FluidContainer>
  <router-view v-else-if="repo && repoPermissions" />
</template>

<script lang="ts">
import { computed, defineComponent, onMounted, provide, ref, toRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Tab from '~/components/tabs/Tab.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import useApiClient from '~/compositions/useApiClient';
import useNotifications from '~/compositions/useNotifications';
import { RepoPermissions } from '~/lib/api/types';
import BuildStore from '~/store/builds';
import RepoStore from '~/store/repos';

export default defineComponent({
  name: 'RepoWrapper',

  components: { FluidContainer, IconButton, Icon, Tabs, Tab },

  props: {
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
  },

  setup(props) {
    const repoOwner = toRef(props, 'repoOwner');
    const repoName = toRef(props, 'repoName');
    const repoStore = RepoStore();
    const buildStore = BuildStore();
    const apiClient = useApiClient();
    const notifications = useNotifications();
    const route = useRoute();
    const router = useRouter();

    const repo = repoStore.getRepo(repoOwner, repoName);
    const repoPermissions = ref<RepoPermissions>();
    const builds = buildStore.getSortedBuilds(repoOwner, repoName);
    provide('repo', repo);
    provide('repo-permissions', repoPermissions);
    provide('builds', builds);

    async function loadRepo() {
      repoPermissions.value = await apiClient.getRepoPermissions(repoOwner.value, repoName.value);
      if (!repoPermissions.value.pull) {
        notifications.notify({ type: 'error', title: 'Not allowed to access this repository' });
        await router.replace({ name: 'home' });
        return;
      }

      await repoStore.loadRepo(repoOwner.value, repoName.value);
      await buildStore.loadBuilds(repoOwner.value, repoName.value);
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

    return { repo, repoPermissions, badgeUrl, activeTab };
  },
});
</script>
