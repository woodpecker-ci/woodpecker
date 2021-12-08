<template>
  <template v-if="build && repo">
    <FluidContainer class="flex border-b mb-4 items-center dark:border-gray-600 min-w-0">
      <IconButton icon="back" class="flex-shrink-0" @click="$router.back()" />
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
    </FluidContainer>

    <div class="p-0 flex flex-col flex-grow">
      <FluidContainer class="flex text-gray-500 justify-between py-0">
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
      </FluidContainer>

      <div v-if="build.status === 'blocked'" class="flex flex-col flex-grow justify-center items-center">
        <Icon name="status-blocked" class="w-32 h-32 text-gray-500" />
        <p class="text-xl text-gray-500">This pipeline is awaiting approval by some maintainer!</p>
        <div v-if="repoPermissions.push" class="flex mt-2 space-x-4">
          <Button color="green" text="Approve" :is-loading="isApprovingBuild" @click="approveBuild" />
          <Button color="red" text="Decline" :is-loading="isDecliningBuild" @click="declineBuild" />
        </div>
      </div>
      <div v-else-if="build.status === 'declined'" class="flex flex-col flex-grow justify-center items-center">
        <Icon name="status-blocked" class="w-32 h-32 text-gray-500" />
        <p class="text-xl text-gray-500">This pipeline has been declined!</p>
      </div>
      <BuildProcs v-else v-model:selected-proc-id="selectedProcId" :build="build" />
    </div>
  </template>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, PropType, Ref, toRef, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import IconButton from '~/components/atomic/IconButton.vue';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import BuildProcs from '~/components/repo/build/BuildProcs.vue';
import BuildStatusIcon from '~/components/repo/build/BuildStatusIcon.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useBuild from '~/compositions/useBuild';
import useNotifications from '~/compositions/useNotifications';
import { Repo, RepoPermissions } from '~/lib/api/types';
import BuildStore from '~/store/builds';
import { findProc } from '~/utils/helpers';

export default defineComponent({
  name: 'RepoBuild',

  components: {
    FluidContainer,
    Button,
    BuildStatusIcon,
    BuildProcs,
    IconButton,
    Icon,
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
    const { since, duration, message } = useBuild(build);
    const procId = toRef(props, 'procId');
    const selectedProcId = computed({
      get() {
        if (procId.value) {
          return parseInt(procId.value, 10);
        }

        if (!build.value || !build.value.procs || !build.value.procs[0].children) {
          return null;
        }

        return build.value.procs[0].children[0].pid;
      },
      set(_selectedProcId: number | null) {
        if (!_selectedProcId) {
          return;
        }

        router.replace({ params: { ...route.params, procId: `${_selectedProcId}` } });
      },
    });

    async function loadBuild(): Promise<void> {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await buildStore.loadBuild(repo.value.owner, repo.value.name, parseInt(buildId.value, 10));
    }

    const { doSubmit: cancelBuild, isLoading: isCancelingBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      if (!build.value.procs) {
        throw new Error('Unexpected: Build procs not loaded');
      }

      // TODO: is selectedProcId right?
      const proc = findProc(build.value.procs, selectedProcId.value || 2);

      if (!proc) {
        throw new Error('Unexpected: Proc not found');
      }

      await apiClient.cancelBuild(repo.value.owner, repo.value.name, parseInt(buildId.value, 10), proc.ppid);
      notifications.notify({ title: 'Pipeline canceled', type: 'success' });
    });

    const { doSubmit: approveBuild, isLoading: isApprovingBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.approveBuild(repo.value.owner, repo.value.name, buildId.value);
      notifications.notify({ title: 'Pipeline approved', type: 'success' });
    });

    const { doSubmit: declineBuild, isLoading: isDecliningBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.declineBuild(repo.value.owner, repo.value.name, buildId.value);
      notifications.notify({ title: 'Pipeline declined', type: 'success' });
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

    return {
      repoPermissions,
      selectedProcId,
      build,
      since,
      duration,
      repo,
      message,
      isCancelingBuild,
      isApprovingBuild,
      isDecliningBuild,
      isRestartingBuild,
      cancelBuild,
      restartBuild,
      approveBuild,
      declineBuild,
    };
  },
});
</script>
