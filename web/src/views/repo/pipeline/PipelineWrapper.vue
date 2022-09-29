<template>
  <template v-if="pipeline && repo">
    <FluidContainer class="flex flex-col min-w-0 dark:border-gray-600">
      <div class="flex mb-2 items-center <md:flex-wrap">
        <IconButton icon="back" class="flex-shrink-0" @click="goBack" />

        <h1
          class="order-3 w-full <md:flex-wrap md:order-none md:w-auto md:ml-2 flex text-center text-xl text-color whitespace-nowrap overflow-hidden overflow-ellipsis"
        >
          <span class="w-full md:w-auto text-center">{{ $t('repo.pipeline.pipeline', { pipelineId }) }}</span>
          <span class="<md:hidden mx-2">-</span>
          <span class="w-full md:w-auto text-center truncate">{{ message }}</span>
        </h1>

        <PipelineStatusIcon :pipeline="pipeline" class="flex flex-shrink-0 ml-auto" />

        <template v-if="repoPermissions.push">
          <Button
            v-if="pipeline.status === 'pending' || pipeline.status === 'running'"
            class="ml-4 flex-shrink-0"
            :text="$t('repo.pipeline.actions.cancel')"
            :is-loading="isCancelingPipeline"
            @click="cancelPipeline"
          />
          <Button
            v-else-if="pipeline.status !== 'blocked' && pipeline.status !== 'declined'"
            class="ml-4 flex-shrink-0"
            :text="$t('repo.pipeline.actions.restart')"
            :is-loading="isRestartingPipeline"
            @click="restartPipeline"
          />
        </template>
      </div>

      <div class="flex flex-wrap gap-y-2 items-center justify-between">
        <Tabs v-model="activeTab" disable-hash-mode class="order-2 md:order-none">
          <Tab id="tasks" :title="$t('repo.pipeline.tasks')" />
          <Tab id="config" :title="$t('repo.pipeline.config')" />
          <Tab
            v-if="pipeline.event === 'push' || pipeline.event === 'pull_request'"
            id="changed-files"
            :title="$t('repo.pipeline.files', { files: pipeline.changed_files?.length || 0 })"
          />
        </Tabs>

        <div class="flex justify-between gap-x-4 text-color flex-shrink-0 pb-2 md:p-0 mx-auto md:mr-0">
          <div class="flex space-x-1 items-center flex-shrink-0">
            <Icon name="since" />
            <Tooltip>
              <span>{{ since }}</span>
              <template #popper
                ><span class="font-bold">{{ $t('repo.pipeline.created') }}</span> {{ created }}</template
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
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import Tab from '~/components/tabs/Tab.vue';
import Tabs from '~/components/tabs/Tabs.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useFavicon } from '~/compositions/useFavicon';
import useNotifications from '~/compositions/useNotifications';
import usePipeline from '~/compositions/usePipeline';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';
import { Repo, RepoPermissions } from '~/lib/api/types';
import PipelineStore from '~/store/pipelines';

export default defineComponent({
  name: 'PipelineWrapper',

  components: {
    FluidContainer,
    Button,
    PipelineStatusIcon,
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

    pipelineId: {
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

    const pipelineStore = PipelineStore();
    const pipelineId = toRef(props, 'pipelineId');
    const repoOwner = toRef(props, 'repoOwner');
    const repoName = toRef(props, 'repoName');
    const repo = inject<Ref<Repo>>('repo');
    const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
    if (!repo || !repoPermissions) {
      throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
    }

    const pipeline = pipelineStore.getPipeline(repoOwner, repoName, pipelineId);
    const { since, duration, created } = usePipeline(pipeline);
    provide('pipeline', pipeline);

    const { message } = usePipeline(pipeline);

    async function loadPipeline(): Promise<void> {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await pipelineStore.loadPipeline(repo.value.owner, repo.value.name, parseInt(pipelineId.value, 10));

      favicon.updateStatus(pipeline.value.status);
    }

    const { doSubmit: cancelPipeline, isLoading: isCancelingPipeline } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      if (!pipeline.value.procs) {
        throw new Error('Unexpected: Pipeline procs not loaded');
      }

      // TODO: is selectedProcId right?
      // const proc = findProc(pipeline.value.procs, selectedProcId.value || 2);

      // if (!proc) {
      //   throw new Error('Unexpected: Proc not found');
      // }

      await apiClient.cancelPipeline(repo.value.owner, repo.value.name, parseInt(pipelineId.value, 10), 0);
      notifications.notify({ title: i18n.t('repo.pipeline.actions.cancel_success'), type: 'success' });
    });

    const { doSubmit: restartPipeline, isLoading: isRestartingPipeline } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.restartPipeline(repo.value.owner, repo.value.name, pipelineId.value, { fork: true });
      notifications.notify({ title: i18n.t('repo.pipeline.actions.restart_success'), type: 'success' });
      // TODO: directly send to newest pipeline?
      await router.push({ name: 'repo', params: { repoName: repo.value.name, repoOwner: repo.value.owner } });
    });

    onMounted(loadPipeline);
    watch([repo, pipelineId], loadPipeline);
    onBeforeUnmount(() => {
      favicon.updateStatus('default');
    });

    const activeTab = computed({
      get() {
        if (route.name === 'repo-pipeline-changed-files') {
          return 'changed-files';
        }

        if (route.name === 'repo-pipeline-config') {
          return 'config';
        }

        return 'tasks';
      },
      set(tab: string) {
        if (tab === 'tasks') {
          router.replace({ name: 'repo-pipeline' });
        }

        if (tab === 'changed-files') {
          router.replace({ name: 'repo-pipeline-changed-files' });
        }

        if (tab === 'config') {
          router.replace({ name: 'repo-pipeline-config' });
        }
      },
    });

    return {
      repoPermissions,
      pipeline,
      repo,
      message,
      isCancelingPipeline,
      isRestartingPipeline,
      activeTab,
      since,
      duration,
      cancelPipeline,
      restartPipeline,
      goBack: useRouteBackOrDefault({ name: 'repo' }),
      created,
    };
  },
});
</script>
