<template>
  <template v-if="build && repo">
    <FluidContainer class="flex flex-col min-w-0">
      <div class="flex border-b pb-4 items-center dark:border-gray-600">
        <IconButton icon="back" class="flex-shrink-0" @click="goBack" />
        <h1 class="text-xl ml-2 text-gray-500 whitespace-nowrap overflow-hidden overflow-ellipsis">
          Pipeline #{{ buildId }} - {{ message }}
        </h1>
        <BuildStatusIcon :build="build" class="flex flex-shrink-0 ml-auto" />
        <template v-if="repoPermissions.push">
          <Button
            v-if="build.status === 'pending' || build.status === 'running'"
            class="ml-4 flex-shrink-0"
            text="Cancel"
            :is-loading="isCancelingBuild"
            @click="cancelBuild"
          />
          <Button
            v-else-if="build.status !== 'blocked' && build.status !== 'declined'"
            class="ml-4 flex-shrink-0"
            text="Restart"
            :is-loading="isRestartingBuild"
            @click="restartBuild"
          />
        </template>
      </div>

      <div class="flex text-gray-500 justify-between px-2 py-4">
        <div class="flex space-x-2 items-center">
          <div class="flex items-center"><img class="w-6" :src="build.author_avatar" /></div>
          <span>{{ build.author }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <Icon v-if="build.event === 'pull_request'" name="pull_request" />
          <Icon v-else-if="build.event === 'deployment'" name="deployment" />
          <Icon v-else-if="build.event === 'tag'" name="tag" />
          <Icon v-else name="push" />
          <a v-if="build.event === 'pull_request'" class="text-link" :href="build.link_url" target="_blank">{{
            `#${build.ref.replaceAll('refs/pull/', '').replaceAll('/merge', '').replaceAll('/head', '')}`
          }}</a>
          <span v-else>{{ build.branch }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <Icon name="commit" />
          <span v-if="build.event === 'pull_request'">{{ build.commit.slice(0, 10) }}</span>
          <a v-else class="text-link" :href="build.link_url" target="_blank">{{ build.commit.slice(0, 10) }}</a>
        </div>
        <div class="flex space-x-2 items-center">
          <Icon name="since" />
          <span>{{ since }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <Icon name="duration" />
          <span>{{ duration }}</span>
        </div>
      </div>

      <Tabs v-model="activeTab" disable-hash-mode>
        <Tab title="Logs" />
        <Tab title="Config" />
      </Tabs>
    </FluidContainer>

    <router-view />
  </template>
</template>

<script lang="ts">
import {
  computed,
  defineComponent,
  inject,
  onBeforeUnmount,
  onMounted,
  PropType,
  provide,
  Ref,
  toRef,
  watch,
} from 'vue';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import BuildStatusIcon from '~/components/repo/build/BuildStatusIcon.vue';
import Tab from '~/components/tabs/Tab.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useBuild from '~/compositions/useBuild';
import { useFavicon } from '~/compositions/useFavicon';
import useNotifications from '~/compositions/useNotifications';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';
import { Repo, RepoPermissions } from '~/lib/api/types';
import BuildStore from '~/store/builds';

export default defineComponent({
  name: 'BuildWrapper',

  components: {
    FluidContainer,
    Button,
    BuildStatusIcon,
    IconButton,
    Tabs,
    Tab,
  },

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

    buildId: {
      type: String,
      required: true,
    },

    // used by toRef
    // eslint-disable-next-line vue/no-unused-properties
    procId: {
      type: String as PropType<string | null>,
      default: null,
    },
  },

  setup(props) {
    const apiClient = useApiClient();
    const router = useRouter();
    const route = useRoute();
    const notifications = useNotifications();
    const favicon = useFavicon();

    const buildStore = BuildStore();
    const buildId = toRef(props, 'buildId');
    const repoOwner = toRef(props, 'repoOwner');
    const repoName = toRef(props, 'repoName');
    const repo = inject<Ref<Repo>>('repo');
    const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
    if (!repo || !repoPermissions) {
      throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
    }

    const build = buildStore.getBuild(repoOwner, repoName, buildId);
    const { since, duration } = useBuild(build);
    provide('build', build);

    const { message } = useBuild(build);

    async function loadBuild(): Promise<void> {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await buildStore.loadBuild(repo.value.owner, repo.value.name, parseInt(buildId.value, 10));

      favicon.updateStatus(build.value.status);
    }

    const { doSubmit: cancelBuild, isLoading: isCancelingBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      if (!build.value.procs) {
        throw new Error('Unexpected: Build procs not loaded');
      }

      // TODO: is selectedProcId right?
      // const proc = findProc(build.value.procs, selectedProcId.value || 2);

      // if (!proc) {
      //   throw new Error('Unexpected: Proc not found');
      // }

      await apiClient.cancelBuild(repo.value.owner, repo.value.name, parseInt(buildId.value, 10), 0);
      notifications.notify({ title: 'Pipeline canceled', type: 'success' });
    });

    const { doSubmit: restartBuild, isLoading: isRestartingBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.restartBuild(repo.value.owner, repo.value.name, buildId.value, { fork: true });
      notifications.notify({ title: 'Pipeline restarted', type: 'success' });
      // TODO: directly send to newest build?
      await router.push({ name: 'repo', params: { repoName: repo.value.name, repoOwner: repo.value.owner } });
    });

    onMounted(loadBuild);
    watch([repo, buildId], loadBuild);
    onBeforeUnmount(() => {
      favicon.updateStatus('default');
    });

    const activeTab = computed({
      get() {
        if (route.name === 'repo-build') {
          return 'logs';
        }
        return 'config';
      },
      set(tab: string) {
        if (tab === 'config') {
          router.replace({ name: 'repo-build-config' });
        } else {
          router.replace({ name: 'repo-build' });
        }
      },
    });

    return {
      repoPermissions,
      build,
      repo,
      message,
      isCancelingBuild,
      isRestartingBuild,
      activeTab,
      since,
      duration,
      cancelBuild,
      restartBuild,
      goBack: useRouteBackOrDefault({ name: 'repo' }),
    };
  },
});
</script>
