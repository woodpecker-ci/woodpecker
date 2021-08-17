<template>
  <template v-if="build && repo">
    <FluidContainer class="flex border-b mb-4 items-center">
      <IconButton :to="{ name: 'repo' }" icon="back" />
      <h1 class="text-xl ml-2">Build #{{ buildId }} - {{ message }}</h1>
      <BuildStatusIcon :build="build" class="flex ml-auto" />
      <Button
        v-if="build.status === 'pending' || build.status === 'running'"
        class="ml-4"
        text="Cancel"
        @click="cancelBuild"
      />
      <Button v-else class="ml-4" text="Restart" @click="restartBuild" />
    </FluidContainer>

    <div class="p-0 flex flex-col flex-grow">
      <FluidContainer class="flex text-gray-500 justify-between py-0">
        <div class="flex space-x-2 items-center">
          <div class="flex items-center"><img class="w-6" :src="build.author_avatar" /></div>
          <span>{{ build.author }}</span>
        </div>
        <div class="flex space-x-2 items-center">
          <Icon name="commit" />
          <a class="text-link" :href="build.link_url" target="_blank">{{ build.commit.slice(0, 10) }}</a>
        </div>
        <div class="flex space-x-2 items-center">
          <Icon name="branch" />
          <span>{{ build.branch }}</span>
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

      <BuildProcs :build="build" v-model:selected-proc-id="selectedProcId" />
    </div>
  </template>
</template>

<script lang="ts">
import { computed, defineComponent, inject, onMounted, Ref, toRef, watch } from 'vue';
import BuildStore from '~/store/builds';
import { Repo } from '~/lib/api/types';
import FluidContainer from '~/components/layout/FluidContainer.vue';
import Button from '~/components/atomic/Button.vue';
import BuildItem from '~/components/repo/BuildItem.vue';
import BuildStatusIcon from '~/components/repo/BuildStatusIcon.vue';
import useBuild from '~/compositions/useBuild';
import { useRouter, useRoute } from 'vue-router';
import BuildProcs from '~/components/repo/BuildProcs.vue';
import useApiClient from '~/compositions/useApiClient';
import { findProc } from '~/utils/helpers';
import IconButton from '~/components/atomic/IconButton.vue';
import Icon from '~/components/atomic/Icon.vue';
import useNotifications from '~/compositions/useNotifications';

export default defineComponent({
  name: 'Build',

  components: {
    FluidContainer,
    Button,
    BuildItem,
    BuildStatusIcon,
    BuildProcs,
    IconButton,
    Icon,
  },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },
    repoName: {
      type: String,
      required: true,
    },
    buildId: {
      type: String,
      required: true,
    },
    procId: {
      type: String,
      required: false,
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
    if (!repo) {
      throw new Error('Unexpected: "repo" should be provided at this place');
    }

    const build = buildStore.getBuild(repoOwner, repoName, buildId);
    const { since, duration, message } = useBuild(build);
    const procId = toRef(props, 'procId');
    const selectedProcId = computed({
      get() {
        if (procId.value) {
          return parseInt(procId.value);
        }

        if (!build.value || !build.value.procs || !build.value.procs[0].children) {
          return null;
        }

        return build.value.procs[0].children[0].pid;
      },
      set(selectedProcId: number | null) {
        if (!selectedProcId) {
          return;
        }

        router.replace({ params: { ...route.params, procId: selectedProcId } });
      },
    });

    async function loadBuild(): Promise<void> {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await buildStore.loadBuild(repo.value.owner, repo.value.name, parseInt(buildId.value));
    }

    async function cancelBuild() {
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

      await apiClient.cancelBuild(repo.value.owner, repo.value.name, parseInt(buildId.value), proc.ppid);
      notifications.notify({ title: 'Build canceled', type: 'success' });
    }

    // apiClient.approveBuild;
    // apiClient.declineBuild;

    async function restartBuild() {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.restartBuild(repo.value.owner, repo.value.name, buildId.value, { fork: true });
      notifications.notify({ title: 'Build restarted', type: 'success' });
      // TODO: directly send to newest build?
      await router.push({ name: 'repo', params: { repoName: repo.value.name, repoOwner: repo.value.owner } });
    }

    onMounted(loadBuild);
    watch([repo, buildId], loadBuild);

    return { selectedProcId, build, since, duration, repo, message, cancelBuild, restartBuild };
  },
});
</script>
