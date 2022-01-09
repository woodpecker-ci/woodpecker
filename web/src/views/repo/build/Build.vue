<template>
  <div class="p-0 flex flex-col flex-grow">
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

    <div v-else class="flex w-full bg-gray-600 dark:bg-dark-gray-800 min-h-0 flex-grow">
      <div v-if="build.error" class="flex flex-col p-4">
        <span class="text-red-400 font-bold text-xl mb-2">Execution error</span>
        <span class="text-red-400">{{ build.error }}</span>
      </div>

      <BuildProcList v-model:selected-proc-id="selectedProcId" :build="build" />

      <BuildLog v-if="selectedProcId" :build="build" :proc-id="selectedProcId" class="w-9/12 flex-grow" />
    </div>
  </div>
</template>

<script lang="ts">
import { computed, defineComponent, inject, PropType, Ref, toRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Icon from '~/components/atomic/Icon.vue';
import BuildLog from '~/components/repo/build/BuildLog.vue';
import BuildProcList from '~/components/repo/build/BuildProcList.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import useNotifications from '~/compositions/useNotifications';
import { Build, Repo, RepoPermissions } from '~/lib/api/types';

export default defineComponent({
  name: 'Build',

  components: {
    Button,
    BuildProcList,
    Icon,
    BuildLog,
  },

  props: {
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

    const build = inject<Ref<Build>>('build');
    const repo = inject<Ref<Repo>>('repo');
    const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
    if (!repo || !repoPermissions || !build) {
      throw new Error('Unexpected: "repo", "repoPermissions" & "build" should be provided at this place');
    }

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

    const { doSubmit: approveBuild, isLoading: isApprovingBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.approveBuild(repo.value.owner, repo.value.name, `${build.value.number}`);
      notifications.notify({ title: 'Pipeline approved', type: 'success' });
    });

    const { doSubmit: declineBuild, isLoading: isDecliningBuild } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.declineBuild(repo.value.owner, repo.value.name, `${build.value.number}`);
      notifications.notify({ title: 'Pipeline declined', type: 'success' });
    });

    return {
      repoPermissions,
      selectedProcId,
      build,
      isApprovingBuild,
      isDecliningBuild,
      approveBuild,
      declineBuild,
    };
  },
});
</script>
