<template>
  <template v-if="build && repo">
    <FluidContainer class="flex flex-col min-w-0 dark:border-gray-600">
      <div class="flex mb-2 items-center <md:flex-wrap">
        <IconButton icon="back" class="flex-shrink-0" @click="goBack" />

        <h1
          class="
            order-3
            w-full
            <md:flex-wrap
            md:order-none md:w-auto md:ml-2
            flex
            text-center text-xl text-color
            whitespace-nowrap
            overflow-hidden overflow-ellipsis
          "
        >
          <span class="w-full md:w-auto text-center">Pipeline #{{ buildId }}</span>
          <span class="<md:hidden mx-2">-</span>
          <span class="w-full md:w-auto text-center truncate">{{ message }}</span>
        </h1>

        <BuildStatusIcon :build="build" class="flex flex-shrink-0 ml-auto" />

        <template v-if="repoPermissions.push">
          <Button
            v-if="build.status === 'pending' || build.status === 'running'"
            class="ml-4 flex-shrink-0"
            :text="$t('repo.build.actions.cancel')"
            :is-loading="isCancelingBuild"
            @click="cancelBuild"
          />
          <Button
            v-else-if="build.status !== 'blocked' && build.status !== 'declined'"
            class="ml-4 flex-shrink-0"
            :text="$t('repo.build.actions.restart')"
            :is-loading="isRestartingBuild"
            @click="restartBuild"
          />
        </template>
      </div>

      <div class="flex flex-wrap gap-y-2 items-center justify-between">
        <Tabs v-model="activeTab" disable-hash-mode class="order-2 md:order-none">
          <Tab id="tasks" :title="$t('repo.build.tasks')" />
          <Tab id="config" :title="$t('repo.build.config')" />
          <Tab id="changed-files" :title="$t('repo.build.files', { files: build.changed_files?.length || 0 })" />
        </Tabs>

        <div class="flex justify-between gap-x-4 text-color flex-shrink-0 pb-2 md:p-0 mx-auto md:mr-0">
          <div class="flex space-x-1 items-center flex-shrink-0">
            <Icon name="since" />
            <Tooltip>
              <span>{{ since }}</span>
              <template #popper
                ><span class="font-bold">{{ $t('repo.build.created') }}</span> {{ created }}</template
              >
            </Tooltip>
          </div>
          <div class="flex space-x-1 items-center flex-shrink-0">
            <Icon name="duration" />
            <span>{{ duration }}</span>
          </div>
        </div>
      </div>
    </FluidContainer>

    <router-view />
  </template>
</template>

<script lang="ts">
import { Tooltip } from 'floating-vue';
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
import { useI18n } from 'vue-i18n';
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
    Tooltip,
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
    const route = useRoute();
    const router = useRouter();
    const notifications = useNotifications();
    const favicon = useFavicon();
    const i18n = useI18n();

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
    const { since, duration, created } = useBuild(build);
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
      notifications.notify({ title: i18n.t('repo.build.actions.cancel_success'), type: 'success' });
    });

    const { doSubmit: restartBuild, isLoading: isRestartingBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.restartBuild(repo.value.owner, repo.value.name, buildId.value, { fork: true });
      notifications.notify({ title: i18n.t('repo.build.actions.restart_success'), type: 'success' });
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
        if (route.name === 'repo-build-changed-files') {
          return 'changed-files';
        }

        if (route.name === 'repo-build-config') {
          return 'config';
        }

        return 'tasks';
      },
      set(tab: string) {
        if (tab === 'tasks') {
          router.replace({ name: 'repo-build' });
        }

        if (tab === 'changed-files') {
          router.replace({ name: 'repo-build-changed-files' });
        }

        if (tab === 'config') {
          router.replace({ name: 'repo-build-config' });
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
      created,
    };
  },
});
</script>
